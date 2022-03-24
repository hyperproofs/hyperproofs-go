import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import math
import sys
import json

"""
Parse the output to a json and return that as a dataframe.
Will return with all golang's default columns: Time, Action, Package, Output, Elapsed
"""


def get_clean_json_as_df(filename):
    raw = []
    with open(filename, 'r') as f:
        entry = {}
        old_output = ""
        for line in f:
            if line.startswith("{"):
                entry = json.loads(line)
                if "Output" in entry:
                    if line.endswith('\\t"}\n'):
                        old_output += entry["Output"]
                    else:
                        entry["Output"] = old_output + entry["Output"]
                        raw.append(entry)
                        old_output = ""
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
    # df[["Testname", "cores"]] = df["Testname"].str.rsplit(
    #     "-", expand=True)  # Split from the right of the str
    return df


def parse_hashing_benchmarks(df):

    df = get_only_measurements_rows(df)
    df[["Testname", "N"]] = df["Testname"].str.rsplit(
        ";", expand=True)  # Split from the right of the str
    df[["Testname", "Operation"]] = df["Testname"].str.rsplit(
        "/", expand=True)  # Split from the right of the str
    df = df.replace(r'', np.NaN)  # r indicates that it is regular exp

    df = df.reset_index(drop=True)
    df = df[["Operation", "Benchtime", "Time", "N"]]
    return df


def hashing_benchmarks_driver():
    df = get_clean_json_as_df("hashing-benchmark.json",)
    df = parse_hashing_benchmarks(df)
    return df


def parse_micro_macro_benchmarks(df):

    df = get_only_measurements_rows(df)
    df[["Testname", "Txn"]] = df["Testname"].str.rsplit(
        ";", expand=True)  # Split from the right of the str
    df[["Testname", "Ell", "Operation"]] = df["Testname"].str.rsplit(
        "/", expand=True)  # Split from the right of the str
    df["Testname"] = df["Testname"].str.replace("BenchmarkPrunedVCS", "")
    df = df.replace(r'', np.NaN)  # r indicates that it is regular exp

    df = df.reset_index(drop=True)
    df = df[["Operation", "Testname", "Benchtime", "Txn", "Ell", "Time"]]
    df = df.sort_values(by=["Ell", "Txn", "Testname", "Benchtime"])

    return df


def micro_macro_benchmarks_driver(filename):
    df = get_clean_json_as_df(filename)
    df = parse_micro_macro_benchmarks(df)
    return df


if __name__ == '__main__':
    print("Hello, World!")
    # df = hashing_benchmarks_driver()
    # print(df)
    # df = micro_macro_benchmarks_driver("micro-macro-more-runs.json")
    # print(df)
    df = micro_macro_benchmarks_driver("micro-macro-1024txn.json")
    print(df)


    df['Ell'] = df['Ell'].astype(int)
    df = df[df["Ell"] != 10]
    del df['Benchtime']
    df = df.pivot(index=["Operation", "Testname"], columns="Ell")
