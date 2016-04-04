#!/usr/bin/env python2

import argparse
import sys
import time

import common

def main():
    description = ("Launches an MS server and several clients in quick "
                   "succession.")
    parser = argparse.ArgumentParser(description=description)
    parser.add_argument("client_count", type=int, choices=range(2, 7),
                        help="Number of clients to launch")
    args = parser.parse_args()

    print "Starting MS Server (Py)"
    ms_srv_port = 2222
    ms_srv = common.MatchMakingServer(ms_srv_port)
    ms_srv.start()
    time.sleep(1)

    for client_num in range(args.client_count):
        print "Starting client {}".format(client_num + 1)
        node_port = 9999 - (client_num * 3)
        node_rpc_port = 9998 - (client_num * 3)
        http_srv_port = 9997 - (client_num * 3)
        client = common.Client(node_port=node_port,
                               node_rpc_port=node_rpc_port,
                               ms_port=ms_srv_port,
                               http_srv_port=http_srv_port)
        time.sleep(1)
        client.start()

    return "test"

if __name__ == "__main__":
    sys.exit(main())
