#!/usr/bin/env python3
"""Tests of semver.py; run with python3 scripts/semver_test.py."""
import io
import sys
import unittest
from unittest import mock

import semver


class SemverTest(unittest.TestCase):
    def test_valid(self):
        for tag in ["v0.2.0", "v1.4.2", "1.4.2", "v1.5.0-rc.1", "v2.0.0-beta"]:
            self.assertEqual(semver.main(["valid", tag]), 0, tag)
        for tag in ["v1.4", "v01.2.3", "v1.2.3-", "latest", "v1.2.3+build", "v1.2.3 "]:
            self.assertEqual(semver.main(["valid", tag]), 1, tag)

    def test_order(self):
        ordered = ["v0.1.10", "v0.2.0", "v1.0.0-alpha", "v1.0.0-alpha.1", "v1.0.0-alpha.beta",
                   "v1.0.0-beta.2", "v1.0.0-beta.11", "v1.0.0-rc.1", "v1.0.0", "v1.10.0"]
        for lower, higher in zip(ordered, ordered[1:]):
            self.assertEqual(semver.main(["gt", higher, lower]), 0, (higher, lower))
            self.assertEqual(semver.main(["gt", lower, higher]), 1, (lower, higher))
        self.assertEqual(semver.main(["gt", "v1.0.0", "v1.0.0"]), 1)

    def test_max(self):
        stdin = io.StringIO("v0.1.9\nv0.1.10\nv0.2.0-rc.1\nnot-a-version\n")
        out = io.StringIO()
        with mock.patch.object(sys, "stdin", stdin), mock.patch.object(sys, "stdout", out):
            self.assertEqual(semver.main(["max"]), 0)
        self.assertEqual(out.getvalue(), "v0.2.0-rc.1\n")


if __name__ == "__main__":
    unittest.main()
