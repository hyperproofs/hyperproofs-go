import pandas as pd
import sys
import json

"""
Parse the output to a json and return that as a dataframe.
Will return with all golang's default columns: Time, Action, Package, Output, Elapsed
"""


def get_clean_json_as_df(filename):
    raw = []
    with open(filename, 'r') as f:
        for line in f:
            if line.startswith("{"):
                raw.append(json.loads(line))
    df = pd.DataFrame(raw)
    return df


"""
Works only time is measured in "ns/op"
"""


def get_only_measurements_rows(df):
    df = df[["Output"]]
    df = df.dropna()
    # Extract rows with this string
    df = df[df["Output"].str.contains("Benchmark")]
    # Extract only measurement rows
    df = df[df["Output"].str.contains("ns/op")]
    df = df["Output"].str.split(expand=True)
    df.columns = ["Testname", "Benchtime", "Time", "Time_Units",
                  "Memusage", "Memusage_Units", "Mallocs", "Mallocs_Units"]
    df[["Testname", "cores"]] = df["Testname"].str.rsplit(
        "-", expand=True)  # Split from the right of the str
    return df


def parse_hyper_agg_benchmarks(df):

    df = get_only_measurements_rows(df)
    df[["Testname", "Txn"]] = df["Testname"].str.rsplit(
        ";", expand=True)  # Split from the right of the str
    df[["Testname", "Ell", "Operation"]] = df["Testname"].str.rsplit(
        "/", expand=True)  # Split from the right of the str
    df["Operation"] = df["Operation"].str.replace("Aggregate", "")
    df = df.reset_index(drop=True)
    df = df[["Operation", "Txn", "Time"]]

    return df


def hyper_agg_benchmarks_driver(in_filename, out_filename):
    df = get_clean_json_as_df(in_filename)
    df = parse_hyper_agg_benchmarks(df)
    df.to_csv(out_filename, header=None, index=False)
    return df


if __name__ == '__main__':

    in_filename = sys.argv[1]
    out_filename = sys.argv[2]
    print("Reading", in_filename)

    try:
        hyper_agg_benchmarks_driver(
            in_filename, out_filename)
        print("Writing to", out_filename)
    except:
        raise
