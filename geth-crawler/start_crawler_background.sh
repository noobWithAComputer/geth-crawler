#!/bin/bash
./build/bin/geth --syncmode=fast --fakepow --cache=2048 --verbosity=3 2> output.log > error.log &
