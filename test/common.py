#!/usr/bin/env python2

import contextlib
import os
import psutil
import subprocess
import time

_HERE = os.path.dirname(os.path.abspath(__file__))
NODE_CLIENT_DIR = os.path.join(os.path.dirname(_HERE), "Node-Client")
MATCHMAKING_DIR = os.path.join(os.path.dirname(_HERE), "MatchMaking")

@contextlib.contextmanager
def use_cwd(new_cwd):
    original_cwd = os.getcwd()
    os.chdir(new_cwd)
    yield
    os.chdir(original_cwd)

class CommonBinary(object):
    def __init__(self):
        self._process = None
        self._bin_path = None

    def kill(self):
        try:
            psutil.Process(self._process.pid).kill()
        except psutil.NoSuchProcess:
            # This method is best effort, so keep going even if the process
            # can't be killed.
            pass

    def wait(self):
        self._process.wait()

class MatchMakingServer(CommonBinary):
    def __init__(self, port):
        super(MatchMakingServer, self).__init__()
        self._port = port
        self.local_log_filename = "127.0.0.1{}-local.txt".format(port)
        self.govector_log_path = os.path.join(
            MATCHMAKING_DIR, "127.0.0.1{}-Log.txt".format(port))

        possible_bin_paths = [
            os.path.join(MATCHMAKING_DIR, "MS"),
            os.path.join(MATCHMAKING_DIR, "MS.exe"),
        ]
        for possible_bin_path in possible_bin_paths:
            if not os.path.isfile(possible_bin_path):
                continue
            self._bin_path = possible_bin_path
            break

        if not self._bin_path:
            raise Exception("Couldn't find matchmaking binary to run")

    def start(self):
        # We force the working directory to be |MATCHMAKING_DIR| so tests can
        # use a fixed path to log files.
        with use_cwd(MATCHMAKING_DIR), open(os.devnull, "w") as dev_null:
            self._process = subprocess.Popen([self._bin_path,
                                              "localhost:{}".format(self._port)],
                                             stdout=dev_null,
                                             stderr=dev_null)

class Client(CommonBinary):
    def __init__(self, node_port, node_rpc_port, ms_port, http_srv_port):
        super(Client, self).__init__()
        self._node_port = node_port
        self._node_rpc_port = node_rpc_port
        self._ms_port = ms_port
        self._http_srv_port = http_srv_port
        self.local_log_filename = "localhost{}-local.txt".format(node_port)
        # TODO: govector_log_filename should be replaced with govector_log_path
        #       everywhere.
        self.govector_log_filename = "localhost{}-Log.txt".format(node_port)
        self.govector_log_path = os.path.join(
            NODE_CLIENT_DIR, self.govector_log_filename)

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

    def start(self):
        # Our HTML assets are only loaded if we run the binary from the correct
        # cwd.
        with use_cwd(NODE_CLIENT_DIR), open(os.devnull, "w") as dev_null:
            self._process = subprocess.Popen([
                self._bin_path,
                "localhost:{}".format(self._node_port),
                "localhost:{}".format(self._node_rpc_port),
                "localhost:{}".format(self._ms_port),
                "localhost:{}".format(self._http_srv_port)
            ],
            stdout=dev_null,
            stderr=dev_null)

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
        time.sleep(0.5)

    return clients

def kill_remaining_processes():
    """Attempts to kill any remaining client or MS processes."""
    for process in psutil.process_iter():
        process_name = process.name()
        if "Node-Client" in process_name or "MS" in process_name:
            print "Killing stray process '{}'".format(process_name)
            process.kill()
