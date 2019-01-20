#! /usr/bin/env python

import sys
import subprocess
import tempfile
from PIL import Image
import math


start_page = 5
end_page = 29
dpi = 300
pdf_name = sys.argv[1]

border_clr = (186,199,192)
back_clr = (255,255,255)

brown = (140,60,31)
lightbrown = (214, 169, 9)
cyan = (107, 199, 184)
blue = (0, 88, 145)
grey = (159, 161, 164)
green = (81, 149, 53)
purple = (114, 61, 131)
lightblue =(73, 106, 163)
red = (219, 66, 82)
orange = (243, 113, 33)

icons = [
    {"color" : brown, "txt" : "Organico"},
    {"color" : lightbrown, "txt" : "Carta e cartone"},
    {"color" : cyan, "txt" : "Vetro"},
    {"color" : blue, "txt" : "Imballaggi e contenitori"},
    {"color" : grey, "txt" : "Indifferenziata"},
    {"color" : green, "txt" : "Centro di raccolta"},
    {"color" : purple, "txt" : "Ditte specializzate"},
    {"color" : lightblue, "txt" : "Contenitore specifico"},
    {"color" : red, "txt" : "Ritiro ingombranti"},
    {"color" : orange, "txt" : "Rivenditori autorizzati"},
]

row_desc = ["txt", "icons", "icons", "txt"]
FIELD_SEP = ":"

def Dist(a, b):
    sum = 0
    for i in range(len(a)):
        sum += (a[i] - b[i])**2
    return math.sqrt(sum)

def ColorEq(a, b):
    epsilon = 5
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


def Ocr(img):
    tmpfile = tempfile.gettempdir() + "/ocr.png"
    img.save(tmpfile, "PNG")
    tess = ["tesseract", tmpfile, "stdout", "-l", "ita",  "--dpi",  "%d" % dpi]
    return subprocess.check_output(tess)

def findIcons(img):
    msg = set()
    w, h = img.size
    for x in range(w):
        pix = img.getpixel((x, h/2))
        for i in range(len(icons)):
            if ColorEq(pix, icons[i]["color"]):
                msg.add(icons[i]["txt"]) 
                break
    return ", ".join(msg)

def log(str):
    sys.stderr.write(str + "\n")
    sys.stderr.flush()

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
        log("%s: Cannot find table Y start coordinate" % filename)
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
        log("%s: Cannot find table X start coordinate" % filename)
        exit(1)

    col_starts, col_ends = FindColumnsBoundaries(img, 0, y_start)
    log("image: %s" % filename)
    log("x_start %d, y_start %d" % (x_start, y_start))
    log("col_starts, col_ends" +  str(col_starts) + str(col_ends))
    row = 0
    while 1:
        #find row size
        row_size = FindRowSize(img, x_start, y_start)
        if row_size == None:
            return
        log("row %d" % row)
        record = []
        for i in range(len(col_starts)):
            #Crop cell
            left = col_starts[i]
            upper = y_start
            right = col_ends[i]
            lower = y_start + row_size
            im = img.crop((left, upper, right, lower))
            if row_desc[i] == "txt":
                msg = Ocr(im)
                msg = msg.strip().replace("-\n", "").replace("\n", " ")
                record.append(msg)
            elif row_desc[i] == "icons":
                msg = findIcons(im)
                record.append(msg)
            else:
                log("Error in row description format")
                exit(1)
        sys.stdout.write(":".join(record) + "\n")
            
           
        #go to the next row
        for y in range(y_start + row_size, h):    
            pix = img.getpixel((x_start, y))
            if ColorEq(pix, back_clr):
                y_start = y
                break
        else:
            log("%s: Cannot find end of row" % filename)
            exit(1)
        row += 1
    
tmpdir = tempfile.gettempdir()
gs_cmd ="gs -dFirstPage=%d -dLastPage=%d -dNOPAUSE -dBATCH -sDEVICE=png16m -o%s/file-%%03d.png -r%d %s" % (start_page, end_page, tmpdir, dpi, pdf_name)
'''
log("Extracting pages using ghostscript:\n%s\n" % gs_cmd)
ret = subprocess.call(gs_cmd.split())
if ret != 0:
    exit(ret)
'''
for p in range(0, end_page - start_page +1):
    f = "%s/file-%03d.png" % (tmpdir, p+1)
    DecodePage(f)


