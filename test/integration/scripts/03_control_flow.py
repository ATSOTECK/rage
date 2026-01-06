# Test: Control Flow
# Tests if/else, for, while, break, continue

from test_framework import test, expect

def test_if_statements():
    # Simple if
    x = 10
    result = None
    if x > 5:
        result = "greater"
    expect(result).to_be("greater")

    # If-else
    x = 3
    if x > 5:
        result = "greater"
    else:
        result = "not greater"
    expect(result).to_be("not greater")

    # If-elif-else
    x = 5
    if x > 5:
        result = "greater"
    elif x == 5:
        result = "equal"
    else:
        result = "less"
    expect(result).to_be("equal")

    # Nested if
    x = 10
    y = 20
    result = None
    if x > 5:
        if y > 15:
            result = "both"
    expect(result).to_be("both")

def test_for_loops():
    # For loop with range
    total = 0
    for i in range(5):
        total = total + i
    expect(total).to_be(10)

    # For loop with list
    items = [1, 2, 3, 4, 5]
    total = 0
    for item in items:
        total = total + item
    expect(total).to_be(15)

    # For loop with string
    chars = ""
    for c in "hello":
        chars = chars + c
    expect(chars).to_be("hello")

    # Nested for loops
    total = 0
    for i in range(3):
        for j in range(3):
            total = total + 1
    expect(total).to_be(9)

def test_while_loops():
    count = 0
    while count < 5:
        count = count + 1
    expect(count).to_be(5)

def test_break():
    # Break in for loop
    found = -1
    for i in range(10):
        if i == 7:
            found = i
            break
    expect(found).to_be(7)

    # Break in while loop
    count = 0
    while True:
        count = count + 1
        if count >= 5:
            break
    expect(count).to_be(5)

def test_continue():
    evens = []
    for i in range(10):
        if i % 2 != 0:
            continue
        evens.append(i)
    expect(evens).to_be([0, 2, 4, 6, 8])

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
    expect(primes).to_be([2, 3, 5, 7, 11, 13, 17, 19])

test("if_statements", test_if_statements)
test("for_loops", test_for_loops)
test("while_loops", test_while_loops)
test("break", test_break)
test("continue", test_continue)
test("complex_control_flow", test_complex_control_flow)

print("Control flow tests completed")
