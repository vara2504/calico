# Copyright (c) 2015-2016 Tigera, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

def create_bgp_peer(host, scope, ip, asNum):
    assert scope in ('node', 'global')
    node = host.get_hostname() if scope == 'node' else ""
    testdata = {
        'apiVersion': 'v1',
        'kind': 'bgpPeer',
        'metadata': {
            'scope': scope,
            'node': node,
            'peerIP': ip,
        },
        'spec': {
            'asNumber': asNum
        }
    }
    host.writefile("testfile.yaml", testdata)
    host.calicoctl("create -f testfile.yaml")
