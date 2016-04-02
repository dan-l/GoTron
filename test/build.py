#!/usr/bin/env python2

import os
import subprocess
import sys

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
    stages = [
        BuildStage("MS Server",
                   os.path.join(_GOTRON_DIR, "MatchMaking"),
                   ["go", "build", "MS.go", "log.go"]),
        BuildStage("Node Client",
                   os.path.join(_GOTRON_DIR, "Node-Client"),
                   ["go", "build"]),
    ]

    for stage in stages:
        exit_code = stage.run()
        if exit_code != 0:
            return "Stage '{}' failed".format(name)

    print "Successfully executed all build stages"
    return None

if __name__ == "__main__":
    sys.exit(main())
