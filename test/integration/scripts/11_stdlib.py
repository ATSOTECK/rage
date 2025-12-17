# Test: Standard Library Modules
# Tests math, random, string, time, sys

results = {}

# =====================================
# math module
# =====================================
import math

# Constants
results["math_pi_exists"] = math.pi > 3.14
results["math_e_exists"] = math.e > 2.71

# Basic functions
results["math_sqrt"] = math.sqrt(16)
results["math_pow"] = math.pow(2, 10)

# Rounding
results["math_ceil"] = math.ceil(3.2)
results["math_floor"] = math.floor(3.8)

# Other functions
results["math_fabs"] = math.fabs(-5.5)
results["math_factorial"] = math.factorial(5)

# =====================================
# random module
# =====================================
import random

random.seed(42)

# Basic random
r = random.random()
results["random_range"] = 0 <= r < 1

# randint
random.seed(42)
results["random_randint"] = random.randint(1, 10)

# choice
random.seed(42)
results["random_choice"] = random.choice([10, 20, 30, 40, 50])

# shuffle (check it works without error)
random.seed(42)
items = [1, 2, 3, 4, 5]
random.shuffle(items)
results["random_shuffle_len"] = len(items)
results["random_shuffle_sum"] = sum(items)

# =====================================
# string module
# =====================================
import string

results["string_ascii_lower"] = string.ascii_lowercase
results["string_ascii_upper"] = string.ascii_uppercase
results["string_digits"] = string.digits

# =====================================
# time module
# =====================================
import time

# time.time() returns a float
t = time.time()
results["time_time_positive"] = t > 0

# perf_counter
pc1 = time.perf_counter()
pc2 = time.perf_counter()
results["time_perf_counter_order"] = pc2 >= pc1

# monotonic
m1 = time.monotonic()
m2 = time.monotonic()
results["time_monotonic_order"] = m2 >= m1

# =====================================
# sys module
# =====================================
import sys

results["sys_version_exists"] = len(sys.version) > 0
results["sys_platform_exists"] = len(sys.platform) > 0

print("Standard library tests completed")
