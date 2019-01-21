#! /usr/bin/env python
import sys

f = open(sys.argv[1])

for l in f:
    r = l.split(";")
    #print r, len(r)
    assert(len(r)==4)
    assert(r[0] != "")
    if  r[1] == "":
        assert(r[2] != "")
    elif r[2] == "":
        assert(r[1] != "")
    
print "All ok!"