#!/usr/bin/env python2

import contextlib
import multiprocessing
import os
import subprocess

_HERE = os.path.dirname(os.path.abspath(__file__))

@contextlib.contextmanager
def use_cwd(new_cwd):
    original_cwd = os.getcwd()
    os.chdir(new_cwd)
    yield
    os.chdir(original_cwd)

class MatchMakingServer(object):
    def __init__(self, port):
        self.bin_path = None
        self.port = port

        ms_dir = os.path.join(os.path.dirname(_HERE), "MatchMaking")
        possible_bin_paths = [
            os.path.join(ms_dir, "MS"),
            os.path.join(ms_dir, "MS.exe"),
        ]
        for possible_bin_path in possible_bin_paths:
            if not os.path.isfile(possible_bin_path):
                continue
            self.bin_path = possible_bin_path
            break

        if not self.bin_path:
            raise Exception("Couldn't find matchmaking binary to run")

    def _start(self):
        subprocess.call([self.bin_path, "localhost:{}".format(self.port)])

    def start(self):
        ms_server = multiprocessing.Process(target=self._start)
        ms_server.start()

class Client(object):
    def __init__(self, node_port, node_rpc_port, ms_port, http_srv_port):
        self.bin_path = None
        self.node_port = node_port
        self.node_rpc_port = node_rpc_port
        self.ms_port = ms_port
        self.http_srv_port = http_srv_port

        nc_dir = os.path.join(os.path.dirname(_HERE), "Node-Client")
        possible_bin_paths = [
            os.path.join(nc_dir, ".vendor", "bin", "Node-Client"),
            os.path.join(nc_dir, "Node-Client.exe"),
        ]

        env = os.environ
        if "GOBIN" in env:
            possible_bin_paths.append(os.path.join(env["GOBIN"], "Node-Client"))
            possible_bin_paths.append(os.path.join(env["GOBIN"],
                                                   "Node-Client.exe"))

        for possible_bin_path in possible_bin_paths:
            if not os.path.isfile(possible_bin_path):
                continue
            self.bin_path = possible_bin_path
            break

        if not self.bin_path:
            raise Exception("Couldn't find client binary to run")

    def _start(self):
        # Our HTML assets are only loaded if we run the binary from the correct
        # cwd.
        with use_cwd(os.path.join(os.path.dirname(_HERE), "Node-Client")):
            subprocess.call([self.bin_path,
                             "localhost:{}".format(self.node_port),
                             "localhost:{}".format(self.node_rpc_port),
                             "localhost:{}".format(self.ms_port),
                             "localhost:{}".format(self.http_srv_port)])

    def start(self):
        node_client = multiprocessing.Process(target=self._start)
        node_client.start()
