require 'fileutils'
require './build'

ENV['PATH'] = "#{Dir.pwd}/bin:#{ENV['PATH']}"

set_gopath(['.'])

PROTOS = protoc('src/ark')
SRC = FileList['src/ark/**/*'].exclude(/src\/ark\/cmds\/.*/)
DEPS = [:vendor] + SRC + PROTOS

task :atom do
	sh 'atom', '.'
end

task :subl do
	sh 'subl', '.'
end

def commands(paths)
	paths.each do |path|
		name = File.basename(path)
		file path => DEPS + FileList["src/ark/cmds/#{name}/**/*"] do |t|
			sh 'go', 'install', "ark/cmds/#{name}"
		end
	end
end

TARGS = commands([
	'bin/arkd',
	'bin/ark',
])

file 'img/bin/arkd' => DEPS + FileList['src/ark/cmds/arkd/**/*'] do |t|
	FileUtils.mkdir_p('img/bin')
	sh('docker',
		'run',
		'-ti', '--rm',
		'-v', "#{Dir.pwd}/src:/go/src",
		'-v', "#{Dir.pwd}/img/bin:/go/bin",
		'golang:1.6',
		'go', 'install', 'ark/cmds/arkd')
end

file 'bin/govendor' do
	sh 'go', 'get', 'github.com/kardianos/govendor'
	FileUtils.rm_rf('src/github.com')
end

task :vendor => ['bin/govendor'] do
	Dir.chdir('src/ark') do
		sh '../../bin/govendor', 'sync'
		sh '../../bin/govendor', 'install', '+program,vendor'
		sh '../../bin/govendor', 'install', '+vendor'
	end
end

def get_version()
	tag = `git describe`.split('-')
	return tag[0..1].join('')
end

task :dockerize => ['img/bin/arkd'] do
	vers = get_version
	sh 'docker', 'build', '-t', "ark:#{vers}", 'img'
end

task :default => TARGS

task :test do
	sh 'go', 'test', 'ark/web/router', 'ark/store'
end

task :clean do
	TARGS.each do |f|
		FileUtils.rm(f)
	end
end
