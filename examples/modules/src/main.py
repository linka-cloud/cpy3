import sys
from lib.runner import run
from lib import SOME_CONSTANT


def main():
    print(f'running with args: {sys.argv}')
    print(f'constant: {SOME_CONSTANT}')
    run("something")


if __name__ == "__main__":
    main()
