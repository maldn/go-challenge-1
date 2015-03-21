module.exports = (grunt) ->

	grunt.initConfig		
		shell:
			go:
				command: "go test"
		watch:
			test:
				files: [
					"*.go"
				]
				tasks: ["shell:go"]
				options:
					livereload: true
			
		for plugin in [
			'grunt-contrib-watch'
			'grunt-shell']
			grunt.loadNpmTasks plugin

		grunt.registerTask "default", [
			"watch"
		]
