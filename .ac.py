## aut commit ##

import os


TOTAL_COMMITS = 2
COMMIT_MATTER = ""

for x in range(0,TOTAL_COMMITS):
	with open("README1.md","w+") as ff:
		ff.write((COMMIT_MATTER + f" v0.0.9.{x}"))


	os.system("git add . ")
	
	os.system(f"git commit -m \"v0.0.9.{x}\"")

	os.system(f"git status")
