#!/usr/bin/env python2

import argparse
import subprocess
import sys

import common

class BuildStage(object):
    def __init__(self, name, working_dir, command_and_args):
        self.name = name
        self._working_dir = working_dir
        self._command_and_args = command_and_args

    def run(self):
        print "Executing stage '{}'".format(self.name)
        with common.use_cwd(self._working_dir):
            return subprocess.call(self._command_and_args)

def main():
    description = "Quickly builds both the MS and client binaries."
    parser = argparse.ArgumentParser(description=description)
    parser.add_argument("--use-go-build", dest="use_go_build",
                        action="store_true",
                        help="Use |go build| instead of gopm for builing "
                             "the client binary.")
    args = parser.parse_args()

    print "Reminder: Pass the --use-go-build flag if gopm is not being used."

    stages = [
        BuildStage("MS Server",
                   common.MATCHMAKING_DIR,
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
        if stage.run() != 0:
            return "Stage '{}' failed".format(stage.name)

    print "Successfully executed all build stages"
    return None

if __name__ == "__main__":
    sys.exit(main())
