import tempfile
import unittest
from pathlib import Path

import check_readme

RUN = "```sh\ndocker run -d image\n```\n"


class Problems(unittest.TestCase):
    def setUp(self):
        self.dir = tempfile.TemporaryDirectory()
        self.docs = Path(self.dir.name)
        (self.docs / "install").mkdir()
        (self.docs / "install" / "docker.md").write_text("# Docker\n\n" + RUN)
        (self.docs / "guide").mkdir()
        (self.docs / "guide" / "first-backup.md").write_text("# First\n")

    def tearDown(self):
        self.dir.cleanup()

    def test_a_good_readme(self):
        text = f"# R\n\nSee [the guide]({check_readme.SITE}/guide/first-backup) and [COPYING](COPYING).\n\n" + RUN
        self.assertEqual(check_readme.problems(text, self.docs), [])

    def test_what_is_wrong(self):
        text = (
            f"# R\n\n[gone]({check_readme.SITE}/nowhere), ![x](images/missing.png), run `make`.\n\n"
            "```sh\ndocker run -d other\n```\n"
        )
        self.assertEqual(
            check_readme.problems(text, self.docs),
            [
                f"{check_readme.SITE}/nowhere: no page docs/nowhere.md",
                "images/missing.png: no such file",
                "inline code in the prose: `make`",
                "the docker run quick start differs from docs/install/docker.md, which CI runs",
            ],
        )

    def test_code_blocks_are_not_prose(self):
        text = "# R\n\n```sh\necho `date` [x](nowhere.md)\n```\n" + RUN
        self.assertEqual(check_readme.problems(text, self.docs), [])


if __name__ == "__main__":
    unittest.main()
