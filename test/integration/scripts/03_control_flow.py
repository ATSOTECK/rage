# Test: Control Flow
# Tests if/else, for, while, break, continue

results = {}

# Simple if
x = 10
if x > 5:
    results["if_simple"] = "greater"

# If-else
x = 3
if x > 5:
    results["if_else"] = "greater"
else:
    results["if_else"] = "not greater"

# If-elif-else
x = 5
if x > 5:
    results["if_elif"] = "greater"
elif x == 5:
    results["if_elif"] = "equal"
else:
    results["if_elif"] = "less"

# Nested if
x = 10
y = 20
if x > 5:
    if y > 15:
        results["nested_if"] = "both"

# For loop with range
total = 0
for i in range(5):
    total = total + i
results["for_range"] = total

# For loop with list
items = [1, 2, 3, 4, 5]
total = 0
for item in items:
    total = total + item
results["for_list"] = total

# For loop with string
chars = ""
for c in "hello":
    chars = chars + c
results["for_string"] = chars

# Nested for loops
total = 0
for i in range(3):
    for j in range(3):
        total = total + 1
results["nested_for"] = total

# While loop
count = 0
while count < 5:
    count = count + 1
results["while_simple"] = count

# Break in for loop
found = -1
for i in range(10):
    if i == 7:
        found = i
        break
results["break_in_for"] = found

# Continue in for loop
evens = []
for i in range(10):
    if i % 2 != 0:
        continue
    evens.append(i)
results["continue_in_for"] = evens

# Break in while loop
count = 0
while True:
    count = count + 1
    if count >= 5:
        break
results["break_in_while"] = count

# Complex control flow - find primes
primes = []
for num in range(2, 20):
    is_prime = True
    for i in range(2, num):
        if num % i == 0:
            is_prime = False
            break
    if is_prime:
        primes.append(num)
results["find_primes"] = primes

print("Control flow tests completed")
