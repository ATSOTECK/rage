# Benchmark: Estimate Pi using the Leibniz formula
# pi/4 = 1 - 1/3 + 1/5 - 1/7 + 1/9 - ...

def estimate_pi(iterations):
    """Estimate pi using the Leibniz formula."""
    pi_over_4 = 0.0
    sign = 1
    denominator = 1

    i = 0
    while i < iterations:
        pi_over_4 = pi_over_4 + sign / denominator
        sign = -sign
        denominator = denominator + 2
        i = i + 1

    return pi_over_4 * 4

# Run the estimation 100,000 times
iterations_per_run = 10_000
run = 0
total = 0.0

RUNS = 100_000

while run < RUNS:
    pi_estimate = estimate_pi(iterations_per_run)
    # print("Estimate pi to be:", pi_estimate)
    total = total + pi_estimate
    run = run + 1

average_pi = total / RUNS
print("Estimated pi (average of 50 runs):")
print(average_pi)
