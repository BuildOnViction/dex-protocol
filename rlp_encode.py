import sys
import json
from termcolor import colored

def rlp_encode(input):
    if isinstance(input, str):
        if len(input) == 1 and ord(input) < 0x80:
            return input
        else:
            return encode_length(len(input), 0x80) + input
    elif isinstance(input, list):
        output = ''
        for item in input:
            output += rlp_encode(item)
        return encode_length(len(output), 0xc0) + output


def encode_length(L, offset):
    if L < 56:
        return chr(L + offset)
    elif L < 256**8:
        BL = to_binary(L)
        return chr(len(BL) + offset + 55) + BL
    else:
        raise Exception("input too long")


def to_binary(x):
    if x == 0:
        return ''
    else:
        return to_binary(int(x / 256)) + chr(x % 256)


def format_rlp_encode(input):
    output = []
    for c in input:
        ordC = ord(c)
        if ordC >= 0x80:
            output.append("0x{:02x}".format(ordC))
        else:
            output.append("'{}'".format(c))

    return "[ " + ", ".join(output) + " ]"


# run as source file
if __name__ == "__main__":

    represent = "The "
    obj_type = "string"
    if len(sys.argv) == 2:
        input = sys.argv[1]
        if input.startswith("json"):
            input = json.loads(input[4:])
            obj_type = "list"
    else:
        input = sys.argv[1:]
        obj_type = "list"

    if len(input) == 0:
        represent += "empty "
    represent += obj_type

    represent += " {} = ".format(json.dumps(input))
    # finally output
    output = rlp_encode(input)
    represent += format_rlp_encode(output) 
    print(colored(represent, 'green'))
