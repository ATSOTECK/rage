from test_framework import test, expect
import heapq

# =====================
# heapq.heappush / heappop
# =====================

def test_heappush_heappop():
    h = []
    heapq.heappush(h, 5)
    heapq.heappush(h, 3)
    heapq.heappush(h, 7)
    heapq.heappush(h, 1)
    expect(heapq.heappop(h)).to_be(1)
    expect(heapq.heappop(h)).to_be(3)
    expect(heapq.heappop(h)).to_be(5)
    expect(heapq.heappop(h)).to_be(7)

test("heappush/heappop basic ordering", test_heappush_heappop)

def test_heappush_single():
    h = []
    heapq.heappush(h, 42)
    expect(h).to_be([42])
    expect(heapq.heappop(h)).to_be(42)
    expect(h).to_be([])

test("heappush/heappop single element", test_heappush_single)

def test_heappush_duplicates():
    h = []
    heapq.heappush(h, 3)
    heapq.heappush(h, 1)
    heapq.heappush(h, 3)
    heapq.heappush(h, 1)
    expect(heapq.heappop(h)).to_be(1)
    expect(heapq.heappop(h)).to_be(1)
    expect(heapq.heappop(h)).to_be(3)
    expect(heapq.heappop(h)).to_be(3)

test("heappush/heappop with duplicates", test_heappush_duplicates)

# =====================
# heapq.heapify
# =====================

def test_heapify():
    h = [5, 3, 7, 1, 9, 2]
    heapq.heapify(h)
    # After heapify, popping should give sorted order
    result = []
    while h:
        result.append(heapq.heappop(h))
    expect(result).to_be([1, 2, 3, 5, 7, 9])

test("heapify transforms list into heap", test_heapify)

def test_heapify_already_sorted():
    h = [1, 2, 3, 4, 5]
    heapq.heapify(h)
    expect(heapq.heappop(h)).to_be(1)

test("heapify on already sorted list", test_heapify_already_sorted)

def test_heapify_reverse_sorted():
    h = [5, 4, 3, 2, 1]
    heapq.heapify(h)
    result = []
    while h:
        result.append(heapq.heappop(h))
    expect(result).to_be([1, 2, 3, 4, 5])

test("heapify on reverse sorted list", test_heapify_reverse_sorted)

def test_heapify_empty():
    h = []
    heapq.heapify(h)
    expect(h).to_be([])

test("heapify empty list", test_heapify_empty)

# =====================
# heapq.heapreplace
# =====================

def test_heapreplace():
    h = [1, 3, 5, 7]
    heapq.heapify(h)
    result = heapq.heapreplace(h, 4)
    expect(result).to_be(1)
    expect(heapq.heappop(h)).to_be(3)

test("heapreplace pops smallest and pushes new", test_heapreplace)

def test_heapreplace_smaller():
    h = [1, 3, 5]
    heapq.heapify(h)
    # Replacing with something smaller than current min
    result = heapq.heapreplace(h, 0)
    expect(result).to_be(1)
    expect(heapq.heappop(h)).to_be(0)

test("heapreplace with smaller value", test_heapreplace_smaller)

# =====================
# heapq.heappushpop
# =====================

def test_heappushpop():
    h = [1, 3, 5, 7]
    heapq.heapify(h)
    # Push 2 then pop smallest
    result = heapq.heappushpop(h, 2)
    expect(result).to_be(1)

test("heappushpop pushes then pops", test_heappushpop)

def test_heappushpop_smallest():
    h = [3, 5, 7]
    heapq.heapify(h)
    # Push something smaller than everything — it comes right back
    result = heapq.heappushpop(h, 1)
    expect(result).to_be(1)
    expect(heapq.heappop(h)).to_be(3)

test("heappushpop with smallest value returns it", test_heappushpop_smallest)

def test_heappushpop_empty():
    h = []
    result = heapq.heappushpop(h, 5)
    expect(result).to_be(5)
    expect(h).to_be([])

test("heappushpop on empty heap", test_heappushpop_empty)

# =====================
# heapq.nlargest
# =====================

def test_nlargest():
    data = [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]
    result = heapq.nlargest(3, data)
    expect(result).to_be([9, 6, 5])

test("nlargest basic", test_nlargest)

def test_nlargest_all():
    data = [3, 1, 2]
    result = heapq.nlargest(10, data)
    expect(result).to_be([3, 2, 1])

test("nlargest n > len returns all sorted desc", test_nlargest_all)

def test_nlargest_zero():
    result = heapq.nlargest(0, [3, 1, 2])
    expect(result).to_be([])

test("nlargest n=0 returns empty", test_nlargest_zero)

def test_nlargest_with_key():
    data = ["hello", "hi", "hey", "howdy"]
    result = heapq.nlargest(2, data, key=len)
    expect(result).to_be(["howdy", "hello"])

test("nlargest with key function", test_nlargest_with_key)

# =====================
# heapq.nsmallest
# =====================

def test_nsmallest():
    data = [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]
    result = heapq.nsmallest(3, data)
    expect(result).to_be([1, 1, 2])

test("nsmallest basic", test_nsmallest)

def test_nsmallest_all():
    data = [3, 1, 2]
    result = heapq.nsmallest(10, data)
    expect(result).to_be([1, 2, 3])

test("nsmallest n > len returns all sorted asc", test_nsmallest_all)

def test_nsmallest_zero():
    result = heapq.nsmallest(0, [3, 1, 2])
    expect(result).to_be([])

test("nsmallest n=0 returns empty", test_nsmallest_zero)

def test_nsmallest_with_key():
    data = ["hello", "hi", "hey", "howdy"]
    result = heapq.nsmallest(2, data, key=len)
    expect(result).to_be(["hi", "hey"])

test("nsmallest with key function", test_nsmallest_with_key)

# =====================
# heapq.merge
# =====================

def test_merge():
    a = [1, 3, 5, 7]
    b = [2, 4, 6, 8]
    result = heapq.merge(a, b)
    expect(result).to_be([1, 2, 3, 4, 5, 6, 7, 8])

test("merge two sorted lists", test_merge)

def test_merge_three():
    a = [1, 4]
    b = [2, 5]
    c = [3, 6]
    result = heapq.merge(a, b, c)
    expect(result).to_be([1, 2, 3, 4, 5, 6])

test("merge three sorted lists", test_merge_three)

def test_merge_empty():
    result = heapq.merge([], [1, 2], [])
    expect(result).to_be([1, 2])

test("merge with empty lists", test_merge_empty)

def test_merge_reverse():
    a = [5, 3, 1]
    b = [6, 4, 2]
    result = heapq.merge(a, b, reverse=True)
    expect(result).to_be([6, 5, 4, 3, 2, 1])

test("merge with reverse=True", test_merge_reverse)

def test_merge_with_key():
    a = ["a", "ccc"]
    b = ["bb", "dddd"]
    result = heapq.merge(a, b, key=len)
    expect(result).to_be(["a", "bb", "ccc", "dddd"])

test("merge with key function", test_merge_with_key)

# =====================
# Priority queue pattern
# =====================

def test_priority_queue():
    pq = []
    heapq.heappush(pq, (3, "low"))
    heapq.heappush(pq, (1, "high"))
    heapq.heappush(pq, (2, "medium"))
    expect(heapq.heappop(pq)).to_be((1, "high"))
    expect(heapq.heappop(pq)).to_be((2, "medium"))
    expect(heapq.heappop(pq)).to_be((3, "low"))

test("priority queue with tuples", test_priority_queue)

# =====================
# Heap with strings
# =====================

def test_string_heap():
    h = []
    heapq.heappush(h, "banana")
    heapq.heappush(h, "apple")
    heapq.heappush(h, "cherry")
    expect(heapq.heappop(h)).to_be("apple")
    expect(heapq.heappop(h)).to_be("banana")
    expect(heapq.heappop(h)).to_be("cherry")

test("heap with strings", test_string_heap)

# =====================
# Heap with floats
# =====================

def test_float_heap():
    h = [3.14, 2.71, 1.41, 1.73]
    heapq.heapify(h)
    expect(heapq.heappop(h)).to_be(1.41)
    expect(heapq.heappop(h)).to_be(1.73)

test("heap with floats", test_float_heap)

# =====================
# Large heap
# =====================

def test_large_heap():
    data = list(range(100, 0, -1))  # [100, 99, ..., 1]
    heapq.heapify(data)
    result = []
    for i in range(5):
        result.append(heapq.heappop(data))
    expect(result).to_be([1, 2, 3, 4, 5])

test("large heap pop first 5", test_large_heap)
