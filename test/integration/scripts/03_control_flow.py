# Test: Control Flow
# Tests if/else, for, while, break, continue

from test_framework import test, expect

def test_if_statements():
    # Simple if
    x = 10
    result = None
    if x > 5:
        result = "greater"
    expect("greater", result)

    # If-else
    x = 3
    if x > 5:
        result = "greater"
    else:
        result = "not greater"
    expect("not greater", result)

    # If-elif-else
    x = 5
    if x > 5:
        result = "greater"
    elif x == 5:
        result = "equal"
    else:
        result = "less"
    expect("equal", result)

    # Nested if
    x = 10
    y = 20
    result = None
    if x > 5:
        if y > 15:
            result = "both"
    expect("both", result)

def test_for_loops():
    # For loop with range
    total = 0
    for i in range(5):
        total = total + i
    expect(10, total)

    # For loop with list
    items = [1, 2, 3, 4, 5]
    total = 0
    for item in items:
        total = total + item
    expect(15, total)

    # For loop with string
    chars = ""
    for c in "hello":
        chars = chars + c
    expect("hello", chars)

    # Nested for loops
    total = 0
    for i in range(3):
        for j in range(3):
            total = total + 1
    expect(9, total)

def test_while_loops():
    count = 0
    while count < 5:
        count = count + 1
    expect(5, count)

def test_break():
    # Break in for loop
    found = -1
    for i in range(10):
        if i == 7:
            found = i
            break
    expect(7, found)

    # Break in while loop
    count = 0
    while True:
        count = count + 1
        if count >= 5:
            break
    expect(5, count)

def test_continue():
    evens = []
    for i in range(10):
        if i % 2 != 0:
            continue
        evens.append(i)
    expect([0, 2, 4, 6, 8], evens)

def test_complex_control_flow():
    # Find primes
    primes = []
    for num in range(2, 20):
        is_prime = True
        for i in range(2, num):
            if num % i == 0:
                is_prime = False
                break
        if is_prime:
            primes.append(num)
    expect([2, 3, 5, 7, 11, 13, 17, 19], primes)

test("if_statements", test_if_statements)
test("for_loops", test_for_loops)
test("while_loops", test_while_loops)
test("break", test_break)
test("continue", test_continue)
test("complex_control_flow", test_complex_control_flow)

print("Control flow tests completed")
