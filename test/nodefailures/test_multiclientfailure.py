#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class MultiClientFailureTest(unittest.TestCase):
    def tearDown(self):
        common.kill_remaining_processes()

    def test_multi_client_failure(self):
        """A leader client (1) is started followed by two normal clients (2, 3, 4).
        2 and 3 fail. 1 should stay as leader and broadcast that 2 and 3 failed to
        all other clients, and 4 should stay as a client.
        """
        ms_srv_port = 2222
        ms_srv = common.MatchMakingServer(ms_srv_port)
        ms_srv.start()
        time.sleep(2)

        clients = common.start_multiple_clients(ms_srv_port, 4)

        time.sleep(1)

        # Kill client 2 and 3.
        time.sleep(2)
        client2 = clients[1]
        client2.kill()
        client3 = clients[2]
        client3.kill()

        # Wait for (unexpected, but potential) leader re-election to occur.
        time.sleep(8)

        leader = clients[0]
        with open(os.path.join(common.NODE_CLIENT_DIR,
                               leader.local_log_filename)) as log_file:
            found_leader_msg = False
            found_node_msg = False
            for line in log_file:
                if "Im a leader" in line:
                    found_leader_msg = True
                    continue
                elif "Im a node" in line:
                    found_node_msg = True
                    continue
            self.assertFalse(found_node_msg,
                             "Leader should never have been a node")
            self.assertTrue(found_leader_msg,
                            "Leader should always been such")

        client4 = clients[3]
        with open(os.path.join(common.NODE_CLIENT_DIR,
                               client4.local_log_filename)) as log_file:
            found_leader_msg = False
            found_node_msg = False
            for line in log_file:
                if "Im a leader" in line:
                    found_leader_msg = True
                    continue
                elif "Im a node" in line:
                    found_node_msg = True
                    continue
            self.assertTrue(found_node_msg,
                            "Client 4 should have been a node at some point")
            self.assertFalse(found_leader_msg,
                             "Client 4 should never have been a leader")

        # TODO: We should check that the leader informed the other clients that
        #       clients 2 and 3 failed.

        ms_srv.kill()
        ms_srv.wait()
        for client in clients:
            client.kill()
            client.wait()

if __name__ == "__main__":
    unittest.main()
