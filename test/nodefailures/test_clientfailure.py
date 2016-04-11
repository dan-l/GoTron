#!/usr/bin/env python2

import os
import sys
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class ClientFailureTest(common.TestCase):
    def test_client_failure(self):
        """A leader client (1) is started followed by two normal clients (2, 3).
        2 fails, and the leader should inform all other clients that 2 has
        failed. 1 should stay the leader, 3 should stay as a client.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        common.sleep(2)

        clients = common.start_multiple_clients(ms_srv.port, 3)

        common.sleep(common.MatchMakingServer.GAME_START_TIMEOUT)

        # Kill client 2.
        common.sleep(2)
        client2 = clients[1]
        client2.kill()

        # Wait for (unexpected, but potential) leader re-election to occur.
        common.sleep(8)

        leader = clients[0]
        with open(leader.local_log_path) as log_file:
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

        client3 = clients[2]
        with open(client3.local_log_path) as log_file:
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
                            "Client 3 should have been a node at some point")
            self.assertFalse(found_leader_msg,
                             "Client 3 should never have been a leader")

        # TODO: We should check that the leader informed the other clients that
        #       client 2 failed.

if __name__ == "__main__":
    unittest.main()
