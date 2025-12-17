#!/bin/zsh
cloc . --by-file --exclude-dir=.idea,.git --exclude-ext=md,json,mod,sum --not-match-f="scc|cloc|LICENSE" --quiet --report-file=scc.txt
