import os
import sys

ONE_MEBIBYTE = 1048576


def generate_rand_file(filename, size):

    with open(filename, "wb") as f:
        f.write(b"\x00" + os.urandom(size * ONE_MEBIBYTE) + b"\x00")

    print(f"Generated {filename} of size {os.stat(filename).st_size} bytes")


# Usage: python3 filegen.py <file_name> <file_size (in MiB)>
def main():
    n = len(sys.argv)

    if n != 3:
        print("Please specify the file size (MiB).")
        exit(1)

    filename = sys.argv[1]
    filesize = int(sys.argv[2])

    generate_rand_file(filename, filesize)


if __name__ == "__main__":
    main()
