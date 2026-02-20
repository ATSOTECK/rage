from test_framework import test, expect

# Test 1: Basic __matmul__ on a custom class
def test_matmul_basic():
    class Vector:
        def __init__(self, *args):
            self.data = list(args)

        def __matmul__(self, other):
            # Dot product
            result = 0
            for i in range(len(self.data)):
                result = result + self.data[i] * other.data[i]
            return result

    v1 = Vector(1, 2, 3)
    v2 = Vector(4, 5, 6)
    expect(v1 @ v2).to_be(32)  # 1*4 + 2*5 + 3*6

test("basic __matmul__", test_matmul_basic)

# Test 2: __rmatmul__ reverse dispatch
def test_rmatmul():
    class MyObj:
        def __init__(self, value):
            self.value = value

        def __rmatmul__(self, other):
            return other * self.value

    obj = MyObj(10)
    # int doesn't have __matmul__, so MyObj.__rmatmul__ should be called
    result = 3 @ obj
    expect(result).to_be(30)

test("__rmatmul__ reverse dispatch", test_rmatmul)

# Test 3: @= augmented assignment
def test_imatmul():
    class Matrix:
        def __init__(self, value):
            self.value = value

        def __matmul__(self, other):
            return Matrix(self.value * other.value)

    m1 = Matrix(3)
    m2 = Matrix(7)
    m1 @= m2
    expect(m1.value).to_be(21)

test("@= augmented assignment via __matmul__", test_imatmul)

# Test 4: __imatmul__ for in-place operation
def test_imatmul_dunder():
    class Accumulator:
        def __init__(self, value):
            self.value = value

        def __matmul__(self, other):
            return Accumulator(self.value + other.value)

        def __imatmul__(self, other):
            self.value = self.value + other.value
            return self

    a = Accumulator(10)
    a @= Accumulator(5)
    expect(a.value).to_be(15)

test("__imatmul__ in-place operation", test_imatmul_dunder)

# Test 5: Error for unsupported types
def test_matmul_error():
    try:
        result = 3 @ 4
        expect("should have raised").to_be("TypeError")
    except Exception as e:
        expect("unsupported" in str(e)).to_be(True)

test("error for unsupported types", test_matmul_error)
