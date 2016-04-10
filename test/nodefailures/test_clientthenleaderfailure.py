#!/usr/bin/env python2

import os
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class ClientThenLeaderFailureTest(unittest.TestCase):
    def tearDown(self):
        common.kill_remaining_processes()

    def test_client_then_leader_failure(self):
        """A leader client (1) is started followed by three normal clients
        (2, 3, 4). 3 fails followed by the leader client failing. 2 should
        become the new leader and inform all other clients that 3 has failed,
        and 4 should stay as a client.
        """
        ms_srv_port = 2222
        ms_srv = common.MatchMakingServer(ms_srv_port)
        ms_srv.start()
        time.sleep(2)

        clients = common.start_multiple_clients(ms_srv_port, 4)

        time.sleep(2)

        # Kill client 3 and leader.
        client3 = clients[2]
        client3.kill()
        leader = clients[0]
        leader.kill()

        # Wait for leader re-election to occur.
        time.sleep(8)

        client2 = clients[1]
        with open(os.path.join(common.NODE_CLIENT_DIR,
                               client2.local_log_filename)) as log_file:
            found_leader_msg = False
            found_node_msg = False
            found_node_msg_after_leader = False
            for line in log_file:
                if "Im a leader" in line:
                    found_leader_msg = True
                    continue
                elif "Im a node" in line:
                    found_node_msg = True
                    if found_leader_msg:
                        found_node_msg_after_leader = True
                    continue
            self.assertTrue(found_node_msg,
                            "Client 2 should have been a node at some point")
            self.assertTrue(found_leader_msg,
                            "Client 2 should have become the leader")
            self.assertTrue(found_node_msg_after_leader,
                            "Client 2 should not have turn back into a node")

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
                            "Client 4 should have always been a node")
            self.assertFalse(found_leader_msg,
                             "Client 4 should never have been a leader")

        ms_srv.kill()
        ms_srv.wait()
        for client in clients:
            client.kill()
            client.wait()

if __name__ == "__main__":
    unittest.main()
