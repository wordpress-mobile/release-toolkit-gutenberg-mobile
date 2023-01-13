#!/usr/bin/env ruby
require 'erb'
require 'open-uri'
require 'optparse'

class ReleaseChecklist < ERB

  def self.template
    open('https://raw.githubusercontent.com/wordpress-mobile/release-toolkit-gutenberg-mobile/trunk/templates/release_checklist.html.erb').read
  end

  def initialize(version, options = {})
    @version = version
    local_template = options.fetch(:template, false)
    @template = local_template ? File.read(local_template) : self.class.template
    @release_date = options.fetch(:release_date, "[date]")
    @mobile_version = options.fetch(:mobile_version, "XX.X")

    @new_release_url="https://github.com/wordpress-mobile/gutenberg-mobile/releases/new?tag=v#{ @version }&amp;target=release/#{ @version }&amp;title=Release%20<% @version %>"

    super(@template)
  end


  def task_block(description)
  %Q(
  <!-- wp:p2/task {"assigneesList":[]} -->
    <div class="wp-block-p2-task">
      <div>
        <span class="wp-block-p2-task__emoji-status" title="Pending">⬜ </span>
        <div class="wp-block-p2-task__checkbox-wrapper">
          <span title="Pending" class="wp-block-p2-task__checkbox is-disabled is-aria-checked-false"></span>
        </div>
      </div>
      <div class="wp-block-p2-task__main">
        <div class="wp-block-p2-task__left">
          <div class="wp-block-p2-task__content-wrapper">
            <span class="wp-block-p2-task__content">#{description}</span>
          </div>
          <div class="wp-block-p2-task__dates"></div>
        </div>
        <div class="wp-block-p2-task__right">
          <div class="wp-block-p2-task__assignees-avatars"></div>
        </div>
      </div>
    </div>
  <!-- /wp:p2/task -->
  )
  end

  def result
    super(binding)
  end

end

options = {}
option_parser = OptionParser.new do |opts|
  opts.banner = "Usage: release_checklist.rb version [options]"
  opts.on '-d', '--release-date RELEASE_DATE', 'Release date' do |d|
    options[:release_date] = d
  end

  opts.on '-m', '--mobile-version', 'Mobile host version' do |m|
    options[:mobile_version] = m
  end

  opts.on '-t', '--template TEMPLATE', 'Template file' do |t|
    options[:template] = t
  end
end
option_parser.parse!

version = ARGV[0]

unless version.match?(/^\d+\.\d+\.\d+$/)
  STDERR.puts("Valid version is required ( X.XX.X format )")
  exit!
end


STDOUT.puts ReleaseChecklist.new(version, options).result
