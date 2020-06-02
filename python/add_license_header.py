import os
import sys


header = """/*
Copyright (C) 2020 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/"""

directory = sys.argv[1]

for (path, dirnames, filenames) in os.walk(directory):
    print((path, dirnames, filenames))
    for name in filenames:
        fullpath = os.path.join(path, name)
        # print("\t{}".format(fullpath))
        if name[-3:] == ".go":
            with open(fullpath, "r") as infile:
                contents = infile.read()
            if contents[:2] != "/*":
                print("adding license comment to {}".format(fullpath))
                with open(fullpath, "w") as outfile:
                    outfile.write(header + "\n" + contents)
            else:
                print("skipping {}, license comment found".format(fullpath))