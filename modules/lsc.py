#!/usr/bin/env python3
"""
Ls module - a simple module to list directory contents
"""
import json
import sys, os, colorama


def main():
    args = sys.argv[1:]
    
    LS_OUT = {}

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

    NAMES_LIST = []
    SIZE_LIST = []
    EXT_LIST = []
    TYPE_LIST = []

    __DIR = "dir"

    __NAME = ""
    __SIZE = ""
    __EXTENSION = ""

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

                NAMES_LIST.append(str(__NAME))
                SIZE_LIST.append(__SIZE)
                EXT_LIST.append(__EXTENSION)
                TYPE_LIST.append(__DIR)
            else:
                NAMES_LIST.append(str(__NAME))
                #SIZE_LIST.append(__SIZE)
                #EXT_LIST.append(__EXTENSION)
                TYPE_LIST.append(__DIR)
        except:
            continue

    for LIST in [NAMES_LIST, TYPE_LIST, EXT_LIST, SIZE_LIST]:
        if LIST == NAMES_LIST:
            #for xx in LIST:
            #    LS_OUT({"name":xx})
            LS_OUT["name"] = (LIST)

        if LIST == SIZE_LIST:
            #for xx in LIST:
            #    LS_OUT({"size":xx})
            LS_OUT["size"] =(LIST)

        if LIST == EXT_LIST:
            #for xx in LIST:
            #    LS_OUT({"ext":xx})
            LS_OUT["ext"]=(LIST)

        if LIST == TYPE_LIST:
            #for xx in LIST:
            #    LS_OUT({"type":xx})
            LS_OUT["type"]=(LIST)
            ###
        
             

    print(json.dumps({"output": LS_OUT, "status": "success"}))


if __name__ == '__main__':
    main()
