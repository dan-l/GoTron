#!/usr/bin/env python2

import argparse
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

    _ = common.start_multiple_clients(ms_srv_port, args.client_count)

if __name__ == "__main__":
    main()
