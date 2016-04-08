#!/usr/bin/env python2

import contextlib
import multiprocessing
import os
import subprocess
import time

_HERE = os.path.dirname(os.path.abspath(__file__))
NODE_CLIENT_DIR = os.path.join(os.path.dirname(_HERE), "Node-Client")

@contextlib.contextmanager
def use_cwd(new_cwd):
    original_cwd = os.getcwd()
    os.chdir(new_cwd)
    yield
    os.chdir(original_cwd)

class MatchMakingServer(object):
    def __init__(self, port):
        self._bin_path = None
        self._port = port

        ms_dir = os.path.join(os.path.dirname(_HERE), "MatchMaking")
        possible_bin_paths = [
            os.path.join(ms_dir, "MS"),
            os.path.join(ms_dir, "MS.exe"),
        ]
        for possible_bin_path in possible_bin_paths:
            if not os.path.isfile(possible_bin_path):
                continue
            self._bin_path = possible_bin_path
            break

        if not self._bin_path:
            raise Exception("Couldn't find matchmaking binary to run")

    def _start(self):
        subprocess.call([self._bin_path, "localhost:{}".format(self._port)])

    def start(self):
        ms_server = multiprocessing.Process(target=self._start)
        ms_server.start()

class Client(object):
    def __init__(self, node_port, node_rpc_port, ms_port, http_srv_port):
        self._process = None
        self._bin_path = None
        self._node_port = node_port
        self._node_rpc_port = node_rpc_port
        self._ms_port = ms_port
        self._http_srv_port = http_srv_port

        possible_bin_paths = [
            os.path.join(NODE_CLIENT_DIR, ".vendor", "bin", "Node-Client"),
            os.path.join(NODE_CLIENT_DIR, "Node-Client.exe"),
        ]

        env = os.environ
        if "GOBIN" in env:
            possible_bin_paths.append(os.path.join(env["GOBIN"], "Node-Client"))
            possible_bin_paths.append(os.path.join(env["GOBIN"],
                                                   "Node-Client.exe"))

        for possible_bin_path in possible_bin_paths:
            if not os.path.isfile(possible_bin_path):
                print "client not at " + possible_bin_path
                continue
            self._bin_path = possible_bin_path
            print "client is at " + self._bin_path
            break

        if not self._bin_path:
            raise Exception("Couldn't find client binary to run")

    def _start(self):
        # Our HTML assets are only loaded if we run the binary from the correct
        # cwd.
        with use_cwd(os.path.join(os.path.dirname(_HERE), "Node-Client")):
            subprocess.call([self._bin_path,
                             "localhost:{}".format(self._node_port),
                             "localhost:{}".format(self._node_rpc_port),
                             "localhost:{}".format(self._ms_port),
                             "localhost:{}".format(self._http_srv_port)])

    def start(self):
        self._process = multiprocessing.Process(target=self._start)
        self._process.start()

    def kill(self):
        self._process.terminate()

def start_multiple_clients(ms_srv_port, client_count):
    clients = []
    for client_num in range(client_count):
        node_port = 9999 - (client_num * 3)
        node_rpc_port = 9998 - (client_num * 3)
        http_srv_port = 9997 - (client_num * 3)
        clients.append(Client(node_port=node_port,
                              node_rpc_port=node_rpc_port,
                              ms_port=ms_srv_port,
                              http_srv_port=http_srv_port))
        clients[-1].start()
        time.sleep(1)

    return clients
