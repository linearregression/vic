*** Settings ***
Resource  ../../resources/Util.robot

*** Test Cases ***
Test
    ${result}=  Run Process  ${bin-dir}/imagec -standalone -reference photon -destination foo  shell=True  cwd=/
    Should Be Equal As Integers  0  ${result.rc}
    Verify Checksums  /foo/https/registry-1.docker.io/v2/library/photon/latest