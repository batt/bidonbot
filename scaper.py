#! /usr/bin/env python

import sys
import subprocess
import tempfile
from PIL import Image
import math


start_page = 5
end_page = 29
pdf_name = sys.argv[1]

border_clr = (186,199,192)
back_clr = (255,255,255)


def Dist(a, b):
    sum = 0
    for i in range(len(a)):
        sum += (a[i] - b[i])**2
    return math.sqrt(sum)

def ColorEq(a, b):
    epsilon = 1
    return Dist(a, b) < epsilon

def FindRowSize(img, x_start, y_start):
    _, h = img.size
    size = None
    for y in range(y_start, h):
        pix = img.getpixel((x_start, y))
        if ColorEq(pix, border_clr):
            size = y - y_start
            break
    return size

def FindColumnsBoundaries(img, x_start, y_start):
    w, _ = img.size
    col_starts = []
    col_ends = []
    #find columns boundaries
    border = False
    for x in range(x_start, w):
        pix = img.getpixel((x, y_start))
        if border:
            if ColorEq(pix, back_clr):
                border = False
                pix_t = img.getpixel((x, y_start-1))
                if ColorEq(pix_t, border_clr):
                    col_starts.append(x)
        else:
            if ColorEq(pix, border_clr):
                border = True
                pix_t = img.getpixel((x-1, y_start-1))
                if ColorEq(pix_t, border_clr):
                    col_ends.append(x)
    return col_starts, col_ends
    

def DecodePage(filename):
    
    img = Image.open(filename)
    w, h = img.size

    #find upper left corner of the table
    found = 0
    for y in range(h):
        pix = img.getpixel((w/2, y))
        
        if found == 0:
            #first border
            if ColorEq(pix, border_clr):
                found = 1
        elif found == 1:
            #first row (different color from the background)
            if not ColorEq(pix, border_clr):
                found = 2
        elif found == 2:
            #second border
            if ColorEq(pix, border_clr):
                found = 3
        elif found == 3:
            #start of second row (with background color)
            if ColorEq(pix, back_clr):
                y_start = y
                break
    else:
        print "%s: Cannot find table Y start coordinate" % filename
        exit(1)
    
    found = 0
    for x in range(w):
        pix = img.getpixel((x, y_start))
        if found == 0:
            if ColorEq(pix, border_clr):
                found = 1
        elif found == 1:
            if ColorEq(pix, back_clr):
                x_start = x
                break
    else:
        print "%s: Cannot find table X start coordinate" % filename
        exit(1)

    col_starts, col_ends = FindColumnsBoundaries(img, 0, y_start)
    print "image: %s" % filename           
    print "x_start %d, y_start %d" % (x_start, y_start)
    print "col_starts, col_ends", col_starts, col_ends
    row = 0
    while 1:
        #find row size
        row_size = FindRowSize(img, x_start, y_start)
        if row_size == None:
            return
        print "row %d" % row
        for i in range(len(col_starts)):
            #Crop cell
            left = col_starts[i]
            upper = y_start
            right = col_ends[i]
            lower = y_start + row_size
            im = img.crop((left, upper, right, lower))
            #########TODO: extract cell information
            #im.show()
           
        #go to the next row
        for y in range(y_start + row_size, h):    
            pix = img.getpixel((x_start, y))
            if ColorEq(pix, back_clr):
                y_start = y
                break
        else:
            print "%s: Cannot find end of row" % filename
            exit(1)
        row += 1
    
tmpdir = tempfile.gettempdir()
gs_cmd ="gs -dFirstPage=%d -dLastPage=%d -dNOPAUSE -dBATCH -sDEVICE=png16m -o%s/file-%%03d.png -r300 %s" % (start_page, end_page, tmpdir, pdf_name)
'''
print "Extracting pages using ghostscript:\n%s\n" % gs_cmd
ret = subprocess.call(gs_cmd.split())
if ret != 0:
    exit(ret)
'''
for p in range(0, end_page - start_page +1):
    f = "%s/file-%03d.png" % (tmpdir, p+1)
    DecodePage(f)


