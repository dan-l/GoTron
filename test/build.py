#!/usr/bin/env python2

import argparse
import os
import subprocess
import sys

import common

_HERE = os.path.dirname(os.path.abspath(__file__))
_GOTRON_DIR = os.path.dirname(_HERE)

class BuildStage(object):
    def __init__(self, name, working_dir, command_and_args):
        self.name = name
        self.working_dir = working_dir
        self.command_and_args = command_and_args

    def run(self):
        print "Executing stage '{}'".format(self.name)
        os.chdir(self.working_dir)
        return subprocess.call(self.command_and_args)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--use-go-build", dest="use_go_build",
                        action="store_true")
    args = parser.parse_args()

    stages = [
        BuildStage("MS Server",
                   os.path.join(_GOTRON_DIR, "MatchMaking"),
                   ["go", "build", "MS.go", "log.go"]),
    ]

    if args.use_go_build:
        stages.append(BuildStage("Node Client (go build)",
                                 common.NODE_CLIENT_DIR,
                                 ["go", "build"]))
    else:
        stages.append(BuildStage("Node Client (gopm)",
                                 common.NODE_CLIENT_DIR,
                                 ["gopm", "install"]))

    for stage in stages:
        exit_code = stage.run()
        if exit_code != 0:
            return "Stage '{}' failed".format(stage.name)

    print "Successfully executed all build stages"
    return None

if __name__ == "__main__":
    sys.exit(main())
