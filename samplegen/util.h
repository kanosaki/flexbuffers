#pragma once
#include "string"
#include "flatbuffers/flexbuffers.h"
#include <fstream>
#include <iostream>

std::string output_dir;

void gen(std::string label, std::function<void(flexbuffers::Builder &)> fn) {
  flexbuffers::Builder b;
  fn(b);
  std::cout << "Writing: " << label << std::endl;
  b.Finish();
  auto p = output_dir + "/" + label;
  std::fstream out(p, std::fstream::out | std::fstream::binary);
  if (!out) {
    throw std::runtime_error("failed to open file: " +
                             std::string(strerror(errno)));
  }
  auto buf = reinterpret_cast<const char *>(b.GetBuffer().data());
  out.write(buf, b.GetSize());
  out.close();
}

