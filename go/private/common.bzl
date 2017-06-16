# Copyright 2014 The Bazel Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DEFAULT_LIB = "go_default_library"

VENDOR_PREFIX = "/vendor/"

SRC_PREFIX = "/src/" # ADDED BY RECHARGE

go_filetype = FileType([
    ".go",
    ".s",
    ".S",
    ".h",  # may be included by .s
])

# be consistent to cc_library.
hdr_exts = [
    ".h",
    ".hh",
    ".hpp",
    ".hxx",
    ".inc",
]

cc_hdr_filetype = FileType(hdr_exts)

# Extensions of files we can build with the Go compiler or with cc_library.
# This is a subset of the extensions recognized by go/build.
cgo_filetype = FileType([
    ".go",
    ".c",
    ".cc",
    ".cxx",
    ".cpp",
    ".s",
    ".S",
    ".h",
    ".hh",
    ".hpp",
    ".hxx",
])

def get_go_toolchain(ctx):
    return ctx.attr._go_toolchain #TODO(toolchains): ctx.toolchains[go_toolchain_type]

def emit_generate_params_action(cmds, ctx, fn):
  cmds_all = [
      # Use bash explicitly. /bin/sh is default, and it may be linked to a
      # different shell, e.g., /bin/dash on Ubuntu.
      "#!/bin/bash",
      "set -e",
  ]
  cmds_all += cmds
  cmds_all_str = "\n".join(cmds_all) + "\n"
  f = ctx.new_file(ctx.configuration.bin_dir, fn)
  ctx.file_action(
      output = f,
      content = cmds_all_str,
      executable = True)
  return f

