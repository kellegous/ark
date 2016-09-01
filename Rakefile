require 'fileutils'
require './build'

# ENV['PATH'] = "#{Dir.pwd}/bin:#{ENV['PATH']}"

set_gopath(['.'])

GODEPS = go_get('src', [
	'github.com/golang/glog',
])

SRC = FileList['src/dinghy/**/*'].exclude(/src\/dinghy\/cmds\/.*/)
DEPS = GODEPS + SRC

TARGS = [
  'bin/dinghyd',
]

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

commands(TARGS)

task :default => TARGS

task :test do
	sh 'go', 'test', 'dinghy/web/router'
end

task :clean do
	TARGS.each do |f|
		FileUtils.rm(f)
	end
end
