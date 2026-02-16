#!/usr/bin/env python3
"""
Ls module - a simple module to list directory contents
"""
import json
import sys, os, colorama


def main():
    args = sys.argv[1:]
    
    LS_OUT = []

    DL_LIST_DOT_FILES = False
    DL_LIST_ITEM_INFO = False

    dir = "./"

    for ARG_V in args:
        if ARG_V.startswith("-"):
            for item in ARG_V.split("-"):
                if "a" in item:
                    DL_LIST_DOT_FILES = True
                if "l" in item:
                    DL_LIST_ITEM_INFO = True
                if "p" in item:
                    dir = item.replace("p:","")
                else:
                    pass
        else:
            dir = ARG_V

    FILES = os.listdir(dir)

    for EACH in FILES:
        try: 
            __DIR = "dir"

            __NAME = ""
            __SIZE = ""
            __EXTENSION = ""

            ###
            __NAME = EACH

            if os.path.isdir(EACH):
                __DIR = "dir"
            else:
                __DIR = "file"

            if DL_LIST_ITEM_INFO == True:
               
                try:
                    __SIZE = os.path.getsize(EACH)
                    __EXTENSION = str(EACH.split(".")[1:][0])
                except:
                    __SIZE = "??"
                    __EXTENSION = "??"

                LS_OUT.append({
                    "name":__NAME,
                    "size": __SIZE,
                    "ext":__EXTENSION,
                    "type":__DIR
                })
            else:
                LS_OUT.append({
                    "name":__NAME,
                    "type":__DIR
                })
        except:
            continue
        
             

    print(json.dumps({"output": LS_OUT, "status": "success"}))


if __name__ == '__main__':
    main()
