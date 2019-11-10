
# CMakeのバージョンを設定
cmake_minimum_required(VERSION 2.8)
project(samplegen CXX)
set(CMAKE_CXX_STANDARD 17)
add_executable(flexbuffers-samplegen samplegen.cpp util.h)
add_executable(flexbuffers-corpusgen corpusgen.cpp util.h)

set(FLATBUFFERS_BUILD_TESTS Off CACHE INTERNAL "" FORCE)
set(FLATBUFFERS_BUILD_FLATHASH Off CACHE INTERNAL "" FORCE)
set(FLATBUFFERS_INSTALL Off CACHE INTERNAL "" FORCE)

add_subdirectory(${CMAKE_SOURCE_DIR}/third_party/flatbuffers)
add_subdirectory(${CMAKE_SOURCE_DIR}/third_party/rapidjson)

target_link_libraries(flexbuffers-samplegen flatbuffers)
target_link_libraries(flexbuffers-corpusgen flatbuffers)
