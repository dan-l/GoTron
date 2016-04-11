#!/usr/bin/env python2

import argparse
import signal

import common

def main():
    description = ("Launches an MS server and several clients in quick "
                   "succession.")
    parser = argparse.ArgumentParser(description=description)
    parser.add_argument("client_count", type=int, choices=range(2, 7),
                        help="Number of clients to launch")
    args = parser.parse_args()

    def sigint_handler(signum, frame):
        _ = signum, frame
        common.kill_remaining_processes()

    # Catch Ctrl-C interrupts so we can (try to) kill our spawned processes.
    signal.signal(signal.SIGINT, sigint_handler)

    ms_srv = common.MatchMakingServer(2222)
    ms_srv.start()
    common.sleep(2)

    clients = common.start_multiple_clients(ms_srv.port, args.client_count)

    ms_srv.wait()
    for client in clients:
        client.wait()

if __name__ == "__main__":
    main()
