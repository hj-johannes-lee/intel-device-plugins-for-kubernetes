## This is a generated file, do not edit directly. Edit build/docker/templates/intel-dlb-initcontainer.Dockerfile.in instead.
##
## Copyright 2022 Intel Corporation. All Rights Reserved.
##
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
##
## http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.
###
FROM debian:unstable
COPY demo/dlb-init.sh /usr/bin/
ENTRYPOINT [ "/bin/sh", "/usr/bin/dlb-init.sh"]
LABEL vendor='Intel®'
LABEL version='devel'
LABEL release='1'
LABEL name='intel-dlb-initcontainer'
LABEL summary='Intel® DLB initcontainer for Kubernetes'
LABEL description='Intel DLB initcontainer enables DLB VF devices from each PF, and configures them'
