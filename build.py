#!/usr/bin/env python3

import argparse
import gzip
import hashlib
import json
import os
import shutil
import subprocess
import sys
import time

class Make:
    def __init__(self):
        self.RELEASE = False
        self.VERSION = self._get_version()
        self._commit_id = os.getenv('GITHUB_SHA')
        self.GOOS = self._exec('go env GOOS')
        self.GOARCH = self._exec('go env GOARCH')
        self.apps = [
            {
                'name': 'oneauth',
                'dir': 'cmd/oneauth',
                'build': [
                    'darwin/amd64',
                    'darwin/arm64',
                    'linux/amd64'
                ]
            },
            {
                'name': 'oneauth-server',
                'dir': 'cmd/server',
                'matrix': [
                    {'CGO_ENABLED':'0', 'GOOS':'linux', 'GOARCH':'amd64'},
                ]
            },
            {
                'name': 'oneauth-ssh-test-server',
                'dir': 'cmd/ssh-test-server',
                'matrix': [
                    {'CGO_ENABLED':'0', 'GOOS':'linux', 'GOARCH':'amd64'},
                ]
            }
        ]

    @staticmethod
    def _get_version() -> str:
        # on create tag
        ref_name = os.getenv('GITHUB_REF_NAME')
        if ref_name and ref_name.startswith('v'):
            self.RELEASE = True
            return ref_name

        run_id = os.getenv('GITHUB_RUN_ID')
        if run_id:
            return 'v0.0.' + run_id

        return 'v0.0.' + str(int(time.time()))

    @staticmethod
    def _exec(cmd: str) -> str:
        stream = os.popen(cmd)
        return stream.read().strip()

    @staticmethod
    def clean():
        if os.path.exists('./build'):
            shutil.rmtree('./build')

        os.mkdir('./build')

    def build_app(self, conf: dict) -> dict:
        app_name = conf.get('name')
        app_dir = conf.get('dir')
        
        print(f'-- Building {app_name}...')

        upload_files = {}

        matrix = conf.get('matrix')
        if not matrix:
            upload_files.update(
                self.build_bin(app_name, self.GOOS, self.GOARCH, '1', app_dir)
            )
        else:
            for env in matrix:
                files = self.build_bin(app_name, env.get('GOOS'), env.get('GOARCH'), env.get('CGO_ENABLED'), app_dir)
                upload_files.update(files)

        return upload_files

    def build_bin(self, name: str, goos: str, goarch: str, cgo: str, app_dir: str) -> dict:
        my_env = os.environ.copy()
        my_env['CGO_ENABLED'] = cgo
        my_env['GOOS'] = goos
        my_env['GOARCH'] = goarch

        ld_flags = f'-w -s -X \"github.com/vitalvas/oneauth/internal/buildinfo.Version={self.VERSION}\"'

        if self._commit_id:
            ld_flags += f' -X \"github.com/vitalvas/oneauth/internal/buildinfo.Commit={self._commit_id}\"'

        output = f'./build/{goos}/{goarch}/{name}'

        build_cmd = ['go', 'build', '-ldflags', ld_flags, '-o', output, f'{app_dir}/main.go']

        raw = subprocess.Popen(
            build_cmd,
            env=my_env,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )

        for line in raw.stdout:
            print(line.decode('utf-8').strip())

        is_break = False
        for line in raw.stderr:
            print(line.decode('utf-8').strip())
            is_break = True

        if is_break:
            sys.exit(1)

        upload_files = {}
        upload_files.update(self.create_archive(name, goos, goarch))
        upload_files.update(self.create_manifest(name, goos, goarch))

        _update_manifest = self.create_update_manifest(name)
        ref_name = os.getenv('GITHUB_REF_NAME')
        if ref_name and (ref_name == 'master' or ref_name.startswith('v')):
            upload_files.update(_update_manifest)

        return upload_files

    def create_archive(self, name: str, goos: str, goarch: str) -> str:
        file_name = f'{name}_{goos}_{goarch}.gz'
        with gzip.open(f'build/{file_name}', mode='wb', compresslevel=9) as file_gz:
            with open(f'build/{goos}/{goarch}/{name}', 'rb') as file_in:
                shutil.copyfileobj(file_in, file_gz)

        return {
            file_name: f'{self.VERSION}/{file_name}'
        }

    def get_sha256(self, goos: str, goarch: str, name: str) -> str:
        with open(f'build/{goos}/{goarch}/{name}', 'rb') as file:
            return hashlib.sha256(file.read()).hexdigest()

    def create_manifest(self, name: str, goos: str, goarch: str) -> dict:
        manifest = {
            'name': name,
            'version': self.VERSION,
            'sha256': self.get_sha256(goos, goarch, name)
        }
        if self._commit_id:
            manifest['commit'] = self._commit_id

        file_name = f'{name}_{goos}_{goarch}_manifest.json'

        with open(f'build/{file_name}', 'w') as file:
            json.dump(manifest, file)

        return {
            file_name: f'{self.VERSION}/{file_name}'
        }

    def create_update_manifest(self, name: str) -> dict:
        repo = os.getenv('GITHUB_REPOSITORY', 'vitalvas/oneauth')
        manifest = {
            'name': name,
            'version': self.VERSION,
            'remote_prefix': f'https://github-build-artifacts.vitalvas.dev/{repo}/',
        }

        if self.RELEASE:
            manifest['remote_prefix'] = 'https://oneauth-files.vitalvas.dev/release/'

        file_name = f'{name}_update_manifest.json'

        with open(f'build/{file_name}', 'w') as file:
            json.dump(manifest, file)

        return {
            file_name: file_name
        }

    def upload_files(self, files: dict) -> None:
        print('Uploading files...')
        for src, dst in files.items():
            self.upload_file(src, dst)

    def upload_file(self, src: str, dst: str) -> None:
        repo = os.getenv('GITHUB_REPOSITORY')
        upload_cmd = ['aws', 's3', 'cp', f'build/{file}', f's3://vv-github-build-artifacts/{repo}/{file}']

        if self.RELEASE:
            upload_cmd = ['aws', 's3', 'cp', f'build/{file}', f's3://oneauth-files.vitalvas.dev/release/{file}']

        raw = subprocess.Popen(upload_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        raw.wait()

        for line in raw.stdout:
            print(line.decode('utf-8').strip())

        is_break = False
        for line in raw.stderr:
            print(line.decode('utf-8').strip())
            is_break = True

        if is_break:
            sys.exit(1)

    def build(self, app_name: str) -> None:
        print('Building...')
        print(f'Version: {self.VERSION}')
        print(f'GOOS: {self.GOOS}, GOARCH: {self.GOARCH}')

        self.clean()

        github_actions = os.getenv('GITHUB_ACTIONS')

        upload_files = {}

        for app in self.apps:
            if app_name != 'all' and app_name != app.get('name'):
                continue

            build = app.get('build')
            if build and f'{self.GOOS}/{self.GOARCH}' in build:
                upload_files.update(
                    self.build_app(app)
                )
            elif not build:
                if github_actions and f'{self.GOOS}/{self.GOARCH}' == 'linux/amd64':
                    upload_files.update(
                        self.build_app(app)
                    )
                elif not github_actions:
                    upload_files.update(
                        self.build_app(app)
                    )

        if github_actions:
            print('Upload files:')
            self.upload_files(upload_files)


if __name__ == '__main__':
    make = Make()

    parser = argparse.ArgumentParser()
    parser.add_argument('--app', choices=['all'] + [app.get('name') for app in make.apps], default='all')

    args = parser.parse_args()

    make.build(args.app)
