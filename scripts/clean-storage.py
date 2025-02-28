#!/usr/bin/env python3
import os
from argparse import ArgumentParser
from pathlib import Path

domain_endings = [
    ".com",
    ".org",
    ".net",
    ".int",
    ".edu",
    ".gov",
    ".mil",
    ".arpa",
    ".biz",
    ".info",
    ".name",
    ".pro",
    ".aero",
    ".coop",
    ".museum",
    ".io",
    ".ai",
    ".co",
    ".us",
    ".uk",
    ".cn",
]


def clean_github(basedir: os.PathLike):
    github_path = basedir / "github.com"
    dirs = os.listdir(github_path)
    for d in dirs:
        path = github_path / d
        if os.path.isfile(path):
            print("# meet a abnormal file: {}".format(path))
            print("rm {}".format(path))

        for f in os.listdir(path):
            if f.endswith(".git"):
                if os.path.exists(path / f[:-4]):
                    print("rm -rf {}".format(path / f))
                else:
                    # mv abc.git to abc
                    print("mv {} {}".format(path / f, path / f[:-4]))


def clean_storage(basedir: os.PathLike):
    dirs = os.listdir(basedir)
    after_clean = []
    for d in dirs:
        if d.endswith(tuple(domain_endings)):
            continue
        path = os.path.join(basedir, d)
        if os.path.isdir(path):
            print("rm -rf {}".format(path))
        else:
            after_clean.append(d)
    print("# after clean: ")
    for d in after_clean:
        print("#    {}".format(d))


parser = ArgumentParser(
    description="Clean storage dir, only generate shell commands but not execute"
)
parser.add_argument("-d")
args = parser.parse_args()
basedir = args.d

print("# clean storage root dir")
clean_storage(Path(basedir))
print("# clean storage/github.com")
clean_github(Path(basedir))
