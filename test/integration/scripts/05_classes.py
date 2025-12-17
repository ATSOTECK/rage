# Test: Classes
# Status: NOT IMPLEMENTED
#
# This test is a placeholder documenting class features that need to be implemented.
#
# Current Error: 'object' object is not callable
# The VM fails when trying to instantiate a class with ClassName(args).
#
# Features to implement:
# - Class definition with __init__
# - Instance creation (calling class as constructor)
# - Instance attribute access (self.x)
# - Instance method calls
# - Multiple instances with separate state
# - Simple inheritance (class Child(Parent))
# - Method overriding
# - Calling parent methods
#
# Example code that should work:
#
# class Point:
#     def __init__(self, x, y):
#         self.x = x
#         self.y = y
#
#     def sum(self):
#         return self.x + self.y
#
# p = Point(3, 4)  # <-- This fails: 'object' object is not callable
# results["class_attr_x"] = p.x
# results["class_method"] = p.sum()
#
# class Animal:
#     def __init__(self, name):
#         self.name = name
#
# class Dog(Animal):
#     def speak(self):
#         return "Woof!"
#
# dog = Dog("Buddy")
# results["inheritance_name"] = dog.name
# results["inheritance_speak"] = dog.speak()

results = {}
print("Classes tests skipped - not implemented")
