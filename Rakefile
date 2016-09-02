require 'fileutils'
require './build'

ENV['PATH'] = "#{Dir.pwd}/bin:#{ENV['PATH']}"

set_gopath(['.'])

GODEPS = go_get('src', [
	'github.com/golang/glog',
	'github.com/golang/protobuf/...',
	'github.com/syndtr/goleveldb/leveldb',
])

PROTOS = protoc('src/dinghy')
SRC = FileList['src/dinghy/**/*'].exclude(/src\/dinghy\/cmds\/.*/)
DEPS = GODEPS + SRC + PROTOS

task :atom do
	sh 'atom', '.'
end

task :subl do
	sh 'subl', '.'
end

def commands(paths)
	paths.each do |path|
		name = File.basename(path)
		file path => DEPS + FileList["src/dinghy/cmds/#{name}/**/*"] do |t|
			sh 'go', 'install', "dinghy/cmds/#{name}"
		end
	end
end

TARGS = commands([
	'bin/dinghyd'
])

file 'img/bin/dinghyd' => DEPS + FileList['src/dinghy/cmds/dinghyd/**/*'] do |t|
	FileUtils.mkdir_p('img/bin')
	sh('docker',
		'run',
		'-ti', '--rm',
		'-v', "#{Dir.pwd}/src:/go/src",
		'-v', "#{Dir.pwd}/img/bin:/go/bin",
		'golang:1.6',
		'go', 'install', 'dinghy/cmds/dinghyd')
end

task :dockerize => ['img/bin/dinghyd'] do
	sh 'docker', 'build', '-t', 'dinghy', 'img'
end

task :default => TARGS

task :test do
	sh 'go', 'test', 'dinghy/web/router'
end

task :clean do
	TARGS.each do |f|
		FileUtils.rm(f)
	end
end
