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

    ms_srv = common.MatchMakingServer(2222)
    ms_srv.start()
    time.sleep(1)

    clients = common.start_multiple_clients(ms_srv.port, args.client_count)

    # Wait for the processes to end so that Ctrl-C kills all processes at once.
    ms_srv.wait()
    for client in clients:
        client.wait()

if __name__ == "__main__":
    main()
