import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import math
import sys

rcParams = {'font.size': 35,
            'font.weight': 'bold',
            'font.family': 'sans-serif',
            'text.usetex': True,
            'ps.useafm': True,
            'pdf.use14corefonts': True,
            'font.sans-serif': ['Helvetica'],
            # 'text.latex.preamble': ['\\usepackage{helvet}'],
            'axes.unicode_minus': False,
            'xtick.major.pad': 20,
            'ytick.major.pad': 20,
            }


def rmMathFmt(x, y): return '{}'.format(str(x))

# This function returns the proof generation of the OWWB20 snark.
# The name of the new column will be set to tag.
# OWWB20 code is in seconds.


def extract_prover_data(filename, columns_list, column, tag):
    # None of the output from the OWWB20 gives the header. For uniformity we skip as well.
    df = pd.read_csv(filename, header=None)
    df.columns = columns_list
    # Number of swaps 2000 => 4000 Merkle proof verifications => 1000 TXNs
    df['mrkcount'] = 2 * df['swaps']
    df = df[['mrkcount', column]]
    df = df.rename(columns={column: tag})
    df.index = df['mrkcount']
    df = df[[tag]]
    return df


def merkle_aggregations_prover(files, columns_list):
    df = pd.DataFrame()
    for file in files:
        tempdf = extract_prover_data(
            file[0], columns_list, file[1], file[2])
        df = pd.concat([df, tempdf], axis=1)  # Concat columns
        # print(df)
    return df


def hyperproofs_aggregation(file, columns_list):
    df = pd.read_csv(file[0], header=None)
    df.columns = columns_list
    # Go benchmark measurements are in nanoseconds. Convert to seconds.
    df['time'] = df['time'] / 10**9
    df = df.pivot(index="numleaves", columns="operation", values="time")
    df = df.rename(columns={file[1]: file[2]})
    df = df[[file[2]]]
    # print(df)
    return df


def extrapolate_owwb20(df):

    df = df.drop(columns=['Hyperproofs'])
    # Prover time per Merkle proof
    df_d = df.div(df.index.to_series(), axis=0)
    # Forward fill NA. Other option is fill by mean.
    df_d = df_d.fillna(method='ffill')
    # df_d = df_d.fillna(df_d.mean())
    df_d = df_d.mul(df_d.index.to_series(), axis=0)  # Reconstruct prover time
    df_d = df_d[df.isna()]  # Choose only extrapolated value.
    df_d = df_d.dropna(how='all')  # Drop the rest of the rows
    df = df.loc[df.dropna().index[-1]:]
    df = df.add(df_d, fill_value=0)

    df.index = np.log2(df.index).astype(int)
    df.columns = ["{:} (extrapolated)".format(x) for x in df.columns]
    return df


def agg_prove_plot(df):

    df_c = df[:]
    df_c.index = np.log2(df_c.index).astype(int)
    df_b = extrapolate_owwb20(df[:])

    plt.rcParams.update(rcParams)
    f, ax = plt.subplots(figsize=(12, 9))
    df_c.plot(ax=ax, marker=".", linewidth=6, markersize=20)
    colors = [x.get_color() for x in ax.get_lines()[-len(df_b.columns):]]
    df_b.plot(ax=ax, linestyle="dotted", marker=".", legend=False,
              color=colors, linewidth=6, markersize=20)
    ax.set_yscale('log')
    ax.set_ylabel("Proving time (s)", fontsize=rcParams['font.size'] + 5)
    ax.set_xlabel(
        "Aggregation size ($\mathbf{\log_2}$ scale)", fontsize=rcParams['font.size'] + 5)
    plt.xticks(fontsize=rcParams['font.size'] + 5)
    plt.yticks(fontsize=rcParams['font.size'] + 5)
    ax.xaxis.set_ticks_position('bottom')
    ax.yaxis.set_ticks_position('left')
    ax.tick_params(length=10, which="major", direction='out')
    ax.tick_params(length=5, which="minor", direction='out')
    # Has to be interger. Else, FuncFormatter will truncate the decimal values.
    ax.xaxis.set_ticks(np.arange(2, 15, 2))
    def fmter(x, y): return '$\mathbf{10^{' + str(int(math.log10(x))) + '}}$'
    ax.yaxis.set_major_formatter(plt.FuncFormatter(fmter))
    ax.xaxis.set_major_formatter(
        plt.FuncFormatter(lambda x, y: '{:.0f}'.format(x)))
    plt.grid(True, which="both")
    plt.tight_layout(pad=0.08)
    plt.savefig("aggregation-prover-log.pdf")


def merkle_aggregations_verifier(filename):

    sr = pd.read_json(filename, typ="series")
    df = sr.to_frame()
    df.columns = ["time"]
    df["key"] = df.index
    df = df[["key", "time"]]

    df_tmp = df["key"].str.split(";", expand=True)
    df_tmp.columns = ["N", "M", "operation"]
    df = pd.concat([df, df_tmp], axis=1)
    del df["key"]
    df["N"] = df["N"].astype(int)
    df["M"] = df["M"].astype(int)
    df = df.sort_values(by=["N", "operation", "M"])
    df = df.reset_index(drop=True)
    del df["M"]
    df = df.pivot(index="N", columns="operation", values="time")
    del df["G1MulVecBinaryAvg"]
    del df["G1MulVecRandomAvg"]
    df.index.name = None
    df.columns.name = None
    df = df[["G1MulVecBinary"]]
    df = df.rename(columns={"G1MulVecBinary": "Merkle"})
    # df = df[["G1MulVecRandom"]]
    # df = df.rename(columns={"G1MulVecRandom": "Merkle"})

    df = df / 10**9
    return df


def agg_verify_plot(df):

    df_c = df[:]
    df_c.index = np.log2(df_c.index).astype(int)

    plt.rcParams.update(rcParams)
    f, ax = plt.subplots(figsize=(12, 9))
    df_c.plot(ax=ax, marker=".", linewidth=6, markersize=20)
    ax.set_yscale('log')
    ax.set_ylabel("Verification time (s)", fontsize=rcParams['font.size'] + 5)
    ax.set_xlabel(
        "Aggregation size ($\mathbf{\log_2}$ scale)", fontsize=rcParams['font.size'] + 5)
    plt.xticks(fontsize=rcParams['font.size'] + 5)
    plt.yticks(fontsize=rcParams['font.size'] + 5)
    ax.xaxis.set_ticks_position('bottom')
    ax.yaxis.set_ticks_position('left')
    ax.tick_params(length=10, which="major", direction='out')
    ax.tick_params(length=5, which="minor", direction='out')
    # Has to be interger. Else, FuncFormatter will truncate the decimal values.
    ax.xaxis.set_ticks(np.arange(2, 15, 2))
    def fmter(x, y): return '$\mathbf{10^{' + str(int(math.log10(x))) + '}}$'
    ax.yaxis.set_major_formatter(plt.FuncFormatter(fmter))
    ax.xaxis.set_major_formatter(
        plt.FuncFormatter(lambda x, y: '{:.0f}'.format(x)))
    plt.grid(True, which="both")
    plt.tight_layout(pad=0.08)
    plt.savefig("aggregation-verifier-log.pdf")


def agg_end_to_end_plot(pDf, vDf):

    df_v = vDf
    df_v.index = np.log2(df_v.index).astype(int)

    df_c = pDf[:]
    df_c.index = np.log2(df_c.index).astype(int)
    df_b = extrapolate_owwb20(pDf[:])

    # df_v.to_csv("verification.csv", index_label="numtxn")
    # df_c.to_csv("prover-asis.csv", index_label="numtxn")
    # df_b.to_csv("prover-extra.csv", index_label="numtxn")

    df_c["Hyperproofs"] = df_c["Hyperproofs"] + df_v["Hyperproofs"]
    df_c["Merkle (Poseidon)"] = df_c["Merkle (Poseidon)"] + df_v["Merkle"]
    df_c["Merkle (Pedersen)"] = df_c["Merkle (Pedersen)"] + df_v["Merkle"]

    df_b["Merkle (Poseidon) (extrapolated)"] += df_v["Merkle"]
    df_b["Merkle (Pedersen) (extrapolated)"] += df_v["Merkle"]

    plt.rcParams.update(rcParams)
    f, ax = plt.subplots(figsize=(12, 9))
    df_c.plot(ax=ax, marker=".", linewidth=6, markersize=20)
    colors = [x.get_color() for x in ax.get_lines()[-len(df_b.columns):]]
    df_b.plot(ax=ax, linestyle="dotted", marker=".", legend=False,
              color=colors, linewidth=6, markersize=20)
    # plt.legend(prop={'size': 30})
    ax.set_yscale('log')
    ax.set_ylabel("Proving + Verification time (s)",
                  fontsize=rcParams['font.size'] + 5)
    ax.set_xlabel(
        "Aggregation size ($\mathbf{\log_2}$ scale)", fontsize=rcParams['font.size'] + 5)
    plt.xticks(fontsize=rcParams['font.size'] + 5)
    plt.yticks(fontsize=rcParams['font.size'] + 5)
    ax.xaxis.set_ticks_position('bottom')
    ax.yaxis.set_ticks_position('left')
    ax.tick_params(length=10, which="major", direction='out')
    ax.tick_params(length=5, which="minor", direction='out')
    # Has to be interger. Else, FuncFormatter will truncate the decimal values.
    ax.xaxis.set_ticks(np.arange(2, 15, 2))
    def fmter(x, y): return '$\mathbf{10^{' + str(int(math.log10(x))) + '}}$'
    ax.yaxis.set_major_formatter(plt.FuncFormatter(fmter))
    ax.xaxis.set_major_formatter(
        plt.FuncFormatter(lambda x, y: '{:.0f}'.format(x)))
    plt.grid(True, which="both")
    plt.tight_layout(pad=0.08)
    plt.savefig("aggregation-e2e-log.pdf")


if __name__ == '__main__':
    print("Hello, World!")
    columns_list = ["type", "swaps", "height", "init",
                    "paramgen", "synth", "prover", "verifier"]
    folder = "./"
    files = [
        ("poseidon-30-single.csv", "prover", "Merkle (Poseidon)"),
        ("pedersen-30-single.csv", "prover", "Merkle (Pedersen)")
    ]
    files = [("{}{}".format(folder, x[0]), x[1], x[2]) for x in files]
    df1 = merkle_aggregations_prover(files, columns_list)

    hp_columns_list = ["operation", "numleaves", "time"]
    df2 = hyperproofs_aggregation(("{}{}".format(
        folder, "hyperproofs-agg.csv"), "Prove", "Hyperproofs"), hp_columns_list)

    proveDf = pd.concat([df2, df1], axis=1)
    agg_prove_plot(proveDf)

    df3 = hyperproofs_aggregation(("{}{}".format(
        folder, "hyperproofs-agg.csv"), "Verify", "Hyperproofs"), hp_columns_list)

    df4 = merkle_aggregations_verifier("{}{}".format(
        folder, "benchmarking-snarks-verifier.json"))

    verifyDf = pd.concat([df3, df4], axis=1)
    agg_verify_plot(verifyDf)

    agg_end_to_end_plot(proveDf, verifyDf)
