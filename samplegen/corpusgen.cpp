#include "flatbuffers/flexbuffers.h"
#include "iostream"
#include "util.h"
#include <exception>
#include <iostream>
#include <random>
#include <string>
#include <sys/stat.h>
#include <utility>
#include <vector>

using namespace flexbuffers;

constexpr static const char alphanum[] = "0123456789"
                                         "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
                                         "abcdefghijklmnopqrstuvwxyz";
static const std::vector<uint64_t> num_int_bounds = {INT8_MAX, INT16_MAX,
                                                     INT32_MAX, INT64_MAX};
static const std::vector<uint64_t> num_uint_bounds = {UINT8_MAX, UINT16_MAX,
                                                      UINT32_MAX, UINT64_MAX};
class Generator {
public:
  explicit Generator(unsigned int seed, Builder *bld)
      : engine(seed), leaf_type_dist(0, 15),
        char_dist(0, (sizeof(alphanum) / sizeof(alphanum[0])) - 1),
        max_depth(std::uniform_int_distribution(1, 3)(engine)),
        node_width_dist(0, 3), num_sign_dist(0, 1), b(bld),
        key_max_len(5), string_max_len(5), blob_max_len(5) {}

  void add_leaf(int current_depth) {
    std::string s;
    uint64_t num;
    std::set<std::string> used_key;
    std::vector<uint8_t> blob;
    int width;
    switch (leaf_type_dist(engine)) {
    case 0:
      b->Null();
      break;
    case 1:
      num = gen_num(num_int_bounds);
      b->Int(num_sign_dist(engine) == 0 ? num : -num - 1);
      break;
    case 2:
      b->UInt(gen_num(num_uint_bounds));
      break;
    case 3:
      if (std::uniform_int_distribution(0, 1)(engine) == 0) {
        b->Float(std::uniform_real_distribution<float>()(engine));
      } else {
        b->Double(std::uniform_real_distribution<double>()(engine));
      }
      break;
    case 4:
      s = gen_random_string(1, key_max_len);
      b->Key(s);
      break;
    case 5:
      s = gen_random_string(0, string_max_len);
      b->String(s);
      break;
    case 6:
      num = gen_num(num_int_bounds);
      b->IndirectInt(num_sign_dist(engine) == 0 ? num : -num - 1);
      break;
    case 7:
      b->IndirectUInt(gen_num(num_uint_bounds));
      break;
    case 8:
      if (std::uniform_int_distribution(0, 1)(engine) == 0) {
        b->IndirectFloat(std::uniform_real_distribution<float>()(engine));
      } else {
        b->IndirectDouble(std::uniform_real_distribution<double>()(engine));
      }
      break;
    case 9:
      blob = gen_random_blob(0, blob_max_len);
      b->Blob(blob);
      break;
    case 10:
      b->Bool(std::uniform_int_distribution(0, 1)(engine) == 0);
      break;
    default:
      if (std::uniform_int_distribution(0, 1)(engine) == 0) {
        // Add map
        width = node_width_dist(engine);
        b->Map([&]() {
          if (current_depth > max_depth) {
            return;
          }
          for (int i = 0; i < width; i++) {
            std::string key;
            do {
              key = gen_random_string(1, 10);
            } while (used_key.find(key) != used_key.end());
            used_key.insert(key);
            b->Key(key);
            add_leaf(current_depth + 1);
          }
        });
      } else {
        // Add vector
        width = node_width_dist(engine);
        b->Vector([&]() {
          if (current_depth > max_depth) {
            return;
          }
          for (int i = 0; i < width; i++) {
            add_leaf(current_depth + 1);
          }
        });
      }
      break;
    }
  }

  std::vector<uint8_t> gen_random_blob(int min_len, int max_len) {
    auto len = std::uniform_int_distribution(min_len, max_len)(engine);
    static std::uniform_int_distribution<uint8_t> byte_dist;

    std::vector<uint8_t> buf(len);
    for (int i = 0; i < len; ++i) {
      buf[i] = alphanum[byte_dist(engine)];
    }
    return std::move(buf);
  }

  std::string gen_random_string(int min_len, int max_len) {
    auto key_len = std::uniform_int_distribution(min_len, max_len)(engine);

    std::string s(key_len, ' ');
    for (int i = 0; i < key_len; ++i) {
      s[i] = alphanum[char_dist(engine)];
    }
    return std::move(s);
  }

  uint64_t gen_num(const std::vector<uint64_t> &upper_bounds) {
    std::uniform_int_distribution num_type_dist(0, (int)upper_bounds.size());

    uint64_t upper = upper_bounds[num_type_dist(engine)];
    return std::uniform_int_distribution<uint64_t>(0, upper)(engine);
  }

private:
  std::default_random_engine engine;
  std::uniform_int_distribution<int> leaf_type_dist;
  std::uniform_int_distribution<int> node_width_dist;
  std::uniform_int_distribution<int> char_dist;
  std::uniform_int_distribution<int> num_sign_dist;
  flexbuffers::Builder *b;
  int max_depth, key_max_len, string_max_len, blob_max_len;
};

int main(int argc, char **argv) {
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
  for (auto i = 0; i < 1000; i++) {
    char label[100];
    snprintf(label, sizeof(label), "%d", i);
    gen(std::string(label), [&](Builder &b) {
      Generator g(i + 10, &b);
      g.add_leaf(0);
    });
  }
  return 0;
}