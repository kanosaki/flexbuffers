
#include "flatbuffers/flexbuffers.h"
#include "util.h"
#include <iostream>
#include <sys/stat.h>
#include <utility>
#include <vector>

using namespace flexbuffers;

void generate_files() {
  gen("single_int_1.flexbuf", [&](Builder &b) { b.Int(1); });
  gen("single_uint_1.flexbuf", [&](Builder &b) { b.UInt(1); });
  gen("single_float_1.flexbuf", [&](Builder &b) { b.Float(1.0f); });
  gen("single_double_1.flexbuf", [&](Builder &b) { b.Float(1.0l); });
  gen("single_indirect_int_1.flexbuf", [&](Builder &b) { b.IndirectInt(1); });
  gen("single_indirect_float_1.flexbuf",
      [&](Builder &b) { b.IndirectFloat(1.0f); });
  gen("single_indirect_double_1.flexbuf",
      [&](Builder &b) { b.IndirectFloat(1.0l); });
  gen("simple_string.flexbuf",
      [&](Builder &b) { b.String("hello flexbuffers!"); });
  gen("simple_blob.flexbuf", [&](Builder &b) {
    // no need to manage memory
    auto data = new char[10];
    data[0] = 0;
    data[1] = 3;
    data[2] = 9;
    b.Blob(data, 5);
  });
  gen("simple_map.flexbuf",
      [&](Builder &b) { b.Map([&]() { b.String("foo", "bar"); }); });
  gen("flat_multiple_map.flexbuf", [&](Builder &b) {
    b.Map([&]() {
      b.String("foo", "bar");
      b.Int("a", 123);
      b.Double("b", 12.0);
    });
  });
  gen("simple_vector.flexbuf", [&](Builder &b) {
    b.Vector([&]() {
      b.Add(1);
      b.Add(256);
      b.Add(65546);
    });
  });
  gen("simple_typed_vector.flexbuf", [&](Builder &b) {
    b.TypedVector([&]() {
      b.Add(1);
      b.Add(256);
      b.Add(65546);
    });
  });
  gen("simple_fixed_typed_vector.flexbuf", [&](Builder &b) {
    auto values = new uint32_t[3]{
        1,
        256,
        65546,
    };
    b.FixedTypedVector(values, 3);
  });
  gen("nested_map_vector.flexbuf", [&](Builder &b) {
    b.Map([&]() {
      b.Map("map", [&]() { b.String("foo", "bar"); });
      b.Vector("vec", [&]() {
        b.Add(1);
        b.Add(256);
        b.Add(65546);
      });
      b.Int("int", 123);
    });
  });
  gen("nested_vector_map.flexbuf", [&](Builder &b) {
    b.Vector([&]() {
      b.Map([&]() { b.String("a", "1"); });
      b.Map([&]() { b.Int("b", 1234); });
    });
  });
  gen("primitive_corners.flexbuf", [&](Builder &b) {
    b.Map([&]() {
      b.Int("int32_max", INT32_MAX);
      b.Int("int32_min", INT32_MIN);
      b.Int("int64_max", INT64_MAX);
      b.Int("int64_min", INT64_MIN);
    });
  });
}

int main(int argc, char const *argv[]) {
  if (argc < 2) {
    std::cout << "usage: " << argv[0] << " <output directory> " << std::endl;
    return 1;
  }
  struct stat info {};
  if (stat(argv[1], &info) != 0) {
    std::cout << "directory '" << argv[1] << "' doesn't exist" << std::endl;
    return 1;
  } else if ((info.st_mode & S_IFDIR) == 0) {
    std::cout << argv[1] << " is not a directory." << std::endl;
    return 1;
  }
  output_dir = std::string(argv[1]);
  generate_files();
  return 0;
}
