#!/usr/bin/env python2

import os
import re
import sys
import time
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class BasicFormatTest(unittest.TestCase):
    def tearDown(self):
        common.kill_remaining_processes()

    def test_basic_format(self):
        """Play a game. Communication between nodes, and with the
        matchmaking server should be logged in a file in a ShiViz compatible
        format.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        time.sleep(2)

        clients = common.start_multiple_clients(ms_srv.port, 2)

        # Wait for a bit to make sure a game starts and is played for a while.
        time.sleep(10)

        # This regex is supposed to match lines like:
        #   localhost:9990 {"localhost:9990":2}
        #   127.0.0.1:2222 {"127.0.0.1:2222":2, "localhost:9999":2}
        # Lines in this format are consumable by ShiViz. The () brackets allow
        # the localhost:9990 part above to be separated out.
        govector_log_regex = re.compile(r"(.*:\d{4,5}?) {.*}")
        with open(ms_srv.govector_log_path) as log_file:
            lines = log_file.readlines()
            for line in lines[0::2]:
                self.assertIsNotNone(govector_log_regex.match(line),
                                     "Line '{}' should be ShiViz compat".format(line))
            for line in lines[1::2]:
                self.assertIsNone(govector_log_regex.match(line),
                                  "Line '{}' should be regular text".format(line))

        for client in clients:
            with open(client.govector_log_path) as log_file:
                lines = log_file.readlines()
                for line in lines[0::2]:
                    self.assertIsNotNone(govector_log_regex.match(line),
                                         "Line '{}' should be ShiViz compat".format(line))
                for line in lines[1::2]:
                    self.assertIsNone(govector_log_regex.match(line),
                                      "Line '{}' should be regular text".format(line))

if __name__ == "__main__":
    unittest.main()
