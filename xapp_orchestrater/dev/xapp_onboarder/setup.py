################################################################################
#   Copyright (c) 2020 AT&T Intellectual Property.                             #
#                                                                              #
#   Licensed under the Apache License, Version 2.0 (the "License");            #
#   you may not use this file except in compliance with the License.           #
#   You may obtain a copy of the License at                                    #
#                                                                              #
#       http://www.apache.org/licenses/LICENSE-2.0                             #
#                                                                              #
#   Unless required by applicable law or agreed to in writing, software        #
#   distributed under the License is distributed on an "AS IS" BASIS,          #
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   #
#   See the License for the specific language governing permissions and        #
#   limitations under the License.                                             #
################################################################################

from setuptools import setup, find_packages

with open("README.md", "r") as fh:
    long_description = fh.read()

with open('requirements.txt') as f:
    requirements = f.read().splitlines()


setup(
    name='xapp_onboarder',
    version='1.0.0',
    description='RIC xApp onboarder',
    long_description=long_description,
    long_description_content_type="text/markdown",
    url='https://gerrit.o-ran-sc.org/r/admin/repos/it/dev',
    author='Zhe Huang',
    author_email='zhehuang@research.att.com',
    include_package_data=True,
    packages=find_packages(),
    package_data={'': ['*.yaml', '*.tpl', '*.conf', 'xapp_onboarder', 'cli']},
    classifiers=[
        "Programming Language :: Python :: 3",
        "Operating System :: OS Independent",
    ],
    python_requires='>=3.6',
    install_requires=requirements,
    entry_points={
        'console_scripts': [
            'xapp_onboarder = xapp_onboarder.server.server:main',
            'dms_cli = xapp_onboarder.server.cli:run'
        ]
    },
)
