import os
import re

def extract_go_file_info(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # Extract top-level file description
    description_lines = []
    i = 0
    while i < len(lines) and (lines[i].strip().startswith("//") or lines[i].strip() == ""):
        if lines[i].strip().startswith("//"):
            description_lines.append(lines[i].strip())
        i += 1

    # Extract function declarations and their comments
    func_pattern = re.compile(r'^func\s+.*')
    funcs_with_comments = []

    i = 0
    while i < len(lines):
        if func_pattern.match(lines[i]):
            func_line = lines[i].strip()

            # Collect preceding comment block
            comment_block = []
            j = i - 1
            while j >= 0 and lines[j].strip().startswith("//"):
                comment_block.insert(0, lines[j].strip())
                j -= 1

            funcs_with_comments.append((func_line, comment_block))
        i += 1

    return description_lines, funcs_with_comments

def main():
    for filename in os.listdir('.'):
        if filename.endswith('.go'):
            print(f"\n--- {filename} ---")
            desc, funcs = extract_go_file_info(filename)

            if desc:
                print("File Description:")
                for line in desc:
                    print(" ", line)
            else:
                print("No file-level description found.")

            if funcs:
                print("\nFunctions:")
                for func_line, comments in funcs:
                    if comments:
                        print("  Description:")
                        for line in comments:
                            print("   ", line)
                    else:
                        print("  No description for this function.")
                    print("  Signature:")
                    print("   ", func_line)
                    print()
            else:
                print("No functions found.")

if __name__ == '__main__':
    main()
