load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

"""
Herumi's BLS library for go depends on
- herumi/mcl
- herumi/bls
- sila-chain/bls-sila-go-binary
"""

def bls_dependencies():
    _maybe(
        http_archive,
        name = "sila_bls_sila_go_binary",
        strip_prefix = "bls-sila-go-binary-1.31.1",
        urls = [
            "https://github.com/sila-chain/bls-sila-go-binary/archive/refs/tags/v1.31.1.tar.gz",
        ],
        sha256 = "67827a69cfb50650cacc21f0ad4f3c1dfffa3c81541bc3a1e8b5c1629bf54a46",
        build_file = "@sila//third_party/herumi:bls_sila_go_binary.BUILD",
    )
    _maybe(
        http_archive,
        name = "herumi_mcl",
        strip_prefix = "mcl-0c31ab9648e81f54177325e55ea96dd8e9c8ba6b",
        urls = [
            "https://github.com/herumi/mcl/archive/0c31ab9648e81f54177325e55ea96dd8e9c8ba6b.tar.gz",
        ],
        sha256 = "0be6f61660ad85ab1fdead420f75d59e3ecbf84da7fa1752daf5157c810727c8",
        build_file = "@sila//third_party/herumi:mcl.BUILD",
    )
    _maybe(
        http_archive,
        name = "herumi_bls",
        strip_prefix = "bls-02060e20d81c2714e481922b182b43e8e26d1fee",
        urls = [
            "https://github.com/herumi/bls/archive/02060e20d81c2714e481922b182b43e8e26d1fee.tar.gz",
        ],
        sha256 = "60b405c934514816f5559538dccf95fbdfdcd86ed08bf1fb95daae45f1cabbfd",
        build_file = "@sila//third_party/herumi:bls.BUILD",
    )

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)
