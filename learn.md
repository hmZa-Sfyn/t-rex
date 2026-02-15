# Creating Custom T-Rex Modules 🦖

Learn how to create powerful Python3 modules for the T-Rex shell!

## Module Basics

T-Rex modules are Python3 scripts that:
1. Accept command-line arguments
2. Return **JSON-formatted output** (required!)
3. Follow a simple naming convention

### Requirements

- **Must return valid JSON** - All output must be parseable JSON
- **Standard output** - Use `print()` to output JSON
- **Exit status** - Success (0) or failure (non-zero)
- **Error handling** - Include error messages in JSON when something fails

## Module Template

Here's a minimal module template:

```python
#!/usr/bin/env python3
"""
my_module - Description of what this module does
Usage: my_module [arguments]
"""
import json
import sys

def main():
    # Get command-line arguments
    args = sys.argv[1:]  # Skip program name
    
    # Do your work here
    result = {
        "output": "some result",
        "status": "success"
    }
    
    # Always output JSON
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

## Examples

### Example 1: Simple Echo Module

```python
#!/usr/bin/env python3
"""echo - Repeat back what you say"""
import json
import sys

def main():
    message = " ".join(sys.argv[1:])
    result = {
        "output": message,
        "status": "success"
    }
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

**Usage:**
```bash
❯ echo "hello world"
{"output":"hello world","status":"success"}
```

---

### Example 2: Module with Arguments

```python
#!/usr/bin/env python3
"""calculator - Simple arithmetic operations"""
import json
import sys

def main():
    if len(sys.argv) < 4:
        result = {
            "error": "Usage: calculator <operation> <num1> <num2>",
            "hint": "Operations: add, sub, mul, div"
        }
        print(json.dumps(result))
        return
    
    operation = sys.argv[1]
    try:
        num1 = float(sys.argv[2])
        num2 = float(sys.argv[3])
    except ValueError:
        result = {
            "error": "Invalid numbers provided",
            "hint": "Arguments must be valid numbers"
        }
        print(json.dumps(result))
        return
    
    if operation == "add":
        output = num1 + num2
    elif operation == "sub":
        output = num1 - num2
    elif operation == "mul":
        output = num1 * num2
    elif operation == "div":
        if num2 == 0:
            result = {"error": "Division by zero", "hint": "Second number cannot be 0"}
            print(json.dumps(result))
            return
        output = num1 / num2
    else:
        result = {
            "error": f"Unknown operation: {operation}",
            "hint": "Supported: add, sub, mul, div"
        }
        print(json.dumps(result))
        return
    
    result = {
        "output": output,
        "operation": operation,
        "operands": [num1, num2],
        "status": "success"
    }
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

**Usage:**
```bash
❯ calculator add 5 3
{"output":8,"operation":"add","operands":[5,3],"status":"success"}

❯ calculator div 10 2
{"output":5.0,"operation":"div","operands":[10,2],"status":"success"}
```

---

### Example 3: Working with Lists

```python
#!/usr/bin/env python3
"""files - List files in a directory"""
import json
import sys
import os

def main():
    directory = sys.argv[1] if len(sys.argv) > 1 else "."
    
    if not os.path.isdir(directory):
        result = {
            "error": f"Directory not found: {directory}",
            "hint": "Provide a valid directory path"
        }
        print(json.dumps(result))
        return
    
    files = []
    try:
        for item in os.listdir(directory):
            path = os.path.join(directory, item)
            files.append({
                "name": item,
                "type": "directory" if os.path.isdir(path) else "file",
                "size": os.path.getsize(path)
            })
    except OSError as e:
        result = {
            "error": str(e),
            "hint": "Check directory permissions"
        }
        print(json.dumps(result))
        return
    
    result = {
        "output": files,
        "directory": directory,
        "count": len(files),
        "status": "success"
    }
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

**Usage:**
```bash
❯ files /tmp | pp
output:
  [0]: 
    name: file1.txt
    type: file
    size: 1024
  [1]:
    name: subdir
    type: directory
    size: 4096
```

---

## Module Output Structure

### Success Response

```json
{
  "output": "your result here",
  "status": "success",
  "metadata": "optional additional info"
}
```

### Error Response

```json
{
  "error": "Description of what went wrong",
  "hint": "Suggestion on how to fix it",
  "status": "error"
}
```

### List Response

```json
{
  "output": [
    {"id": 1, "name": "item1"},
    {"id": 2, "name": "item2"}
  ],
  "status": "success"
}
```

## Best Practices

### 1. Always Include Status
```python
result = {
    "output": data,
    "status": "success"  # ✓ Good
}

result = {
    "output": data
}  # ✗ Bad - missing status
```

### 2. Provide Helpful Error Messages
```python
# ✓ Good error
result = {
    "error": "Configuration file not found",
    "hint": "Create ~/.t-rex/.trexrc with proper settings",
    "file": "~/.t-rex/.trexrc"
}

# ✗ Bad error
result = {
    "error": "Error"
}
```

### 3. Handle Edge Cases
```python
# ✓ Handle empty input
if not args:
    result = {
        "error": "No input provided",
        "hint": "Usage: mymodule <argument>"
    }
    print(json.dumps(result))
    return

# ✗ Don't crash on bad input
num = int(sys.argv[1])  # Could crash!
```

### 4. Use Consistent JSON Structure
```python
# ✓ Consistent response format
if success:
    print(json.dumps({"output": result, "status": "success"}))
else:
    print(json.dumps({"error": msg, "hint": help_text}))

# ✗ Inconsistent formats
if success:
    print(json.dumps(result))
else:
    print("ERROR: " + msg)  # Not JSON!
```

## Installing Your Module

### Option 1: Default Modules Directory
```bash
# On Linux/macOS
mkdir -p ~/.t-rex/modules
cp my_module.py ~/.t-rex/modules/

# Now use it
❯ my_module
```

### Option 2: Custom Path
```bash
# Load from custom directory
❯ load -path /path/to/my/modules
❯ my_module
```

### Option 3: Current Directory
```bash
# Put module in current directory
cp my_module.py .
❯ my_module
```

## Piping with Your Modules

Your modules automatically support piping!

```python
#!/usr/bin/env python3
"""users - List users"""
import json

result = {
    "output": [
        {"name": "alice", "uid": 1000, "shell": "/bin/bash"},
        {"name": "bob", "uid": 1001, "shell": "/bin/zsh"}
    ],
    "status": "success"
}
print(json.dumps(result))
```

**Use with piping:**
```bash
❯ users | select name,uid
{"output":[{"name":"alice","uid":1000},{"name":"bob","uid":1001}],"status":"success"}

❯ users | pp
output:
  [0]:
    name: alice
    uid: 1000
    shell: /bin/bash
  [1]:
    name: bob
    uid: 1001
    shell: /bin/zsh
```

## Common Patterns

### Pattern 1: Validation
```python
def validate_input(args):
    if len(args) < required_count:
        return False, "Missing arguments"
    try:
        # Validate types
        int(args[0])
    except ValueError:
        return False, "First argument must be a number"
    return True, None

valid, error = validate_input(args)
if not valid:
    print(json.dumps({"error": error}))
    return
```

### Pattern 2: File Operations
```python
try:
    with open(filename, 'r') as f:
        data = json.load(f)
except FileNotFoundError:
    print(json.dumps({
        "error": f"File not found: {filename}",
        "hint": "Check the file path"
    }))
    return
except json.JSONDecodeError:
    print(json.dumps({
        "error": f"Invalid JSON in: {filename}",
        "hint": "Make sure the file contains valid JSON"
    }))
    return
```

### Pattern 3: External Commands
```python
import subprocess

try:
    result = subprocess.run(['some_command'], 
                          capture_output=True, 
                          text=True,
                          timeout=5)
    output = result.stdout
except subprocess.TimeoutExpired:
    print(json.dumps({
        "error": "Command timed out",
        "hint": "The command took too long to execute"
    }))
except Exception as e:
    print(json.dumps({
        "error": str(e),
        "hint": "Check if the command is installed"
    }))
```

## Testing Your Module

```bash
# Test directly with Python
python3 my_module.py arg1 arg2

# Test with T-Rex
❯ my_module arg1 arg2

# Test piping
❯ my_module | pp
❯ my_module | select field1,field2
```

## Module Tips & Tricks

### Tip 1: Using Environment Variables
```python
import os
module_path = os.getenv('PYTHONPATH', '~/.t-rex/modules')
```

### Tip 2: Writing to stderr for Debug Info
```python
import sys
import json

# Debug info (not output)
print("DEBUG: Processing started", file=sys.stderr)

# Actual output (JSON)
print(json.dumps({"output": result}))
```

### Tip 3: Pretty Printing in Module
```python
# Your module can return already-formatted data
result = {
    "output": {
        "name": "John",
        "age": 30,
        "tags": ["developer", "python"]
    },
    "status": "success"
}
# Users can pipe to 'pp' for pretty printing
```

## Troubleshooting

### Module not found
- Check file is in `~/.t-rex/modules/`
- Ensure filename matches module name
- Module name should be `module_name.py`, not `module_name.txt`

### "Not valid JSON" error
- Module must output ONLY valid JSON to stdout
- No print statements or debug output
- Check for trailing characters

### Wrong output format
- Verify you're using `json.dumps()`
- Check for syntax errors in Python
- Test module directly: `python3 my_module.py`

## Advanced Topics

### Creating Module Packages
```bash
~/.t-rex/modules/
├── math_ops/
│   ├── __init__.py
│   ├── add.py
│   └── multiply.py
└── file_ops/
    ├── __init__.py
    └── list_files.py
```

### Async Modules (advanced)
```python
import asyncio
import json

async def main():
    # Your async code here
    result = await some_async_operation()
    print(json.dumps({"output": result}))

asyncio.run(main())
```

## Next Steps

1. Create your first simple module
2. Test it with `python3 my_module.py`
3. Install it in `~/.t-rex/modules/`
4. Use it in T-Rex shell
5. Combine with piping for powerful commands
6. Share your modules with the community!

Happy module creating! 🦖
