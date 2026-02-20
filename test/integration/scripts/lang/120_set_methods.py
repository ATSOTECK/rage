from test_framework import test, expect

# === difference_update ===
def test_diff_update_basic():
    s = {1, 2, 3, 4, 5}
    s.difference_update({3, 4})
    return s == {1, 2, 5}

test("difference_update basic", lambda: expect(test_diff_update_basic()).to_be(True))

def test_diff_update_list():
    s = {1, 2, 3, 4}
    s.difference_update([2, 4])
    return s == {1, 3}

test("difference_update with list", lambda: expect(test_diff_update_list()).to_be(True))

def test_diff_update_multiple():
    s = {1, 2, 3, 4, 5}
    s.difference_update({1, 2}, {4, 5})
    return s == {3}

test("difference_update multiple args", lambda: expect(test_diff_update_multiple()).to_be(True))

def test_diff_update_empty():
    s = {1, 2, 3}
    s.difference_update(set())
    return s == {1, 2, 3}

test("difference_update with empty", lambda: expect(test_diff_update_empty()).to_be(True))

def test_diff_update_no_args():
    s = {1, 2, 3}
    s.difference_update()
    return s == {1, 2, 3}

test("difference_update no args", lambda: expect(test_diff_update_no_args()).to_be(True))

def test_diff_update_returns_none():
    s = {1, 2}
    result = s.difference_update({1})
    return result is None

test("difference_update returns None", lambda: expect(test_diff_update_returns_none()).to_be(True))

def test_diff_update_all():
    s = {1, 2, 3}
    s.difference_update({1, 2, 3})
    return s == set()

test("difference_update removes all", lambda: expect(test_diff_update_all()).to_be(True))

# === intersection_update ===
def test_inter_update_basic():
    s = {1, 2, 3, 4, 5}
    s.intersection_update({3, 4, 5, 6})
    return s == {3, 4, 5}

test("intersection_update basic", lambda: expect(test_inter_update_basic()).to_be(True))

def test_inter_update_list():
    s = {1, 2, 3, 4}
    s.intersection_update([2, 3])
    return s == {2, 3}

test("intersection_update with list", lambda: expect(test_inter_update_list()).to_be(True))

def test_inter_update_multiple():
    s = {1, 2, 3, 4, 5}
    s.intersection_update({1, 2, 3, 4}, {2, 3, 4, 5})
    return s == {2, 3, 4}

test("intersection_update multiple args", lambda: expect(test_inter_update_multiple()).to_be(True))

def test_inter_update_empty():
    s = {1, 2, 3}
    s.intersection_update(set())
    return s == set()

test("intersection_update with empty", lambda: expect(test_inter_update_empty()).to_be(True))

def test_inter_update_no_args():
    s = {1, 2, 3}
    s.intersection_update()
    return s == {1, 2, 3}

test("intersection_update no args", lambda: expect(test_inter_update_no_args()).to_be(True))

def test_inter_update_returns_none():
    s = {1, 2, 3}
    result = s.intersection_update({2, 3})
    return result is None

test("intersection_update returns None", lambda: expect(test_inter_update_returns_none()).to_be(True))

def test_inter_update_disjoint():
    s = {1, 2, 3}
    s.intersection_update({4, 5, 6})
    return s == set()

test("intersection_update disjoint", lambda: expect(test_inter_update_disjoint()).to_be(True))

# === symmetric_difference_update ===
def test_sym_diff_update_basic():
    s = {1, 2, 3}
    s.symmetric_difference_update({2, 3, 4})
    return s == {1, 4}

test("symmetric_difference_update basic", lambda: expect(test_sym_diff_update_basic()).to_be(True))

def test_sym_diff_update_list():
    s = {1, 2, 3}
    s.symmetric_difference_update([3, 4, 5])
    return s == {1, 2, 4, 5}

test("symmetric_difference_update with list", lambda: expect(test_sym_diff_update_list()).to_be(True))

def test_sym_diff_update_empty():
    s = {1, 2, 3}
    s.symmetric_difference_update(set())
    return s == {1, 2, 3}

test("symmetric_difference_update with empty", lambda: expect(test_sym_diff_update_empty()).to_be(True))

def test_sym_diff_update_same():
    s = {1, 2, 3}
    s.symmetric_difference_update({1, 2, 3})
    return s == set()

test("symmetric_difference_update identical", lambda: expect(test_sym_diff_update_same()).to_be(True))

def test_sym_diff_update_returns_none():
    s = {1, 2}
    result = s.symmetric_difference_update({2, 3})
    return result is None

test("symmetric_difference_update returns None", lambda: expect(test_sym_diff_update_returns_none()).to_be(True))

def test_sym_diff_update_disjoint():
    s = {1, 2}
    s.symmetric_difference_update({3, 4})
    return s == {1, 2, 3, 4}

test("symmetric_difference_update disjoint", lambda: expect(test_sym_diff_update_disjoint()).to_be(True))

# === Verify augmented assignment operators still work ===
def test_ior():
    s = {1, 2}
    s |= {3, 4}
    return s == {1, 2, 3, 4}

test("|= operator", lambda: expect(test_ior()).to_be(True))

def test_iand():
    s = {1, 2, 3, 4}
    s &= {2, 3, 5}
    return s == {2, 3}

test("&= operator", lambda: expect(test_iand()).to_be(True))

def test_isub():
    s = {1, 2, 3, 4}
    s -= {2, 4}
    return s == {1, 3}

test("-= operator", lambda: expect(test_isub()).to_be(True))

def test_ixor():
    s = {1, 2, 3}
    s ^= {2, 3, 4}
    return s == {1, 4}

test("^= operator", lambda: expect(test_ixor()).to_be(True))
