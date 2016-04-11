#!/usr/bin/env python2

import os
import re
import sys
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.dirname(_HERE))

import common

class BasicFormatTest(common.TestCase):
    def test_basic_format(self):
        """Play a game. Communication between nodes, and with the
        matchmaking server should be logged in a file in a ShiViz compatible
        format.
        """
        ms_srv = common.MatchMakingServer(2222)
        ms_srv.start()
        common.sleep(2)

        clients = common.start_multiple_clients(ms_srv.port, 2)

        # Wait for a bit to make sure a game starts and is played for a while.
        common.sleep(common.MatchMakingServer.GAME_START_TIMEOUT * 1.1)

        # This regex is supposed to match lines like:
        #   localhost:9990 {"localhost:9990":2}
        #   127.0.0.1:2222 {"127.0.0.1:2222":2, "localhost:9999":2}
        # Lines in this format are consumable by ShiViz.
        govector_log_regex = re.compile(
            r'(.*:\d{4,5}?) {(".*:\d{4,5}?":\d*)(, ".*:\d{4,5}?":\d*)*}')
        with open(ms_srv.govector_log_path) as log_file:
            lines = log_file.readlines()
            for line in lines[0::2]:
                match = govector_log_regex.match(line)
                self.assertIsNotNone(match,
                                     "Line '{}' should be ShiViz compat".format(line))
                self.assertEquals(match.group(1),
                                  "127.0.0.1:{}".format(ms_srv.port),
                                  "Sender should be the MS")
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
