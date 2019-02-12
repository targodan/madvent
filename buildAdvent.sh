#!/bin/bash

rm -rf open-adventure &>/dev/null

git clone https://gitlab.com/esr/open-adventure.git || exit 1

cd open-adventure || exit 2
patch -i ../open-adventure.patch || exit 3
make || exit 4

cd ..
echo "Done. Executable is ./open-adventure/advent"