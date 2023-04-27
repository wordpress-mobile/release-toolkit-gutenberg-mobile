#!/usr/bin/env ruby
require 'erb'
require 'open-uri'
require 'optparse'
require 'tempfile'

def use_local_template?
  ENV['GBM_USE_LOCAL_RELEASE_TEMPLATES']
end

# Output helpers
def say(message)
  STDERR.print(message)
end

def exit_with(message)
  say(message)
  exit
end

def abort_with(message)
  say(message)
  exit(1)
end

class Checklist < ERB
  def task_block (description)
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
end


class ReleaseChecklist < Checklist

  def self.template
    if use_local_template?
      return File.read('./templates/release_checklist.html.erb')
    end
    open('https://raw.githubusercontent.com/wordpress-mobile/release-toolkit-gutenberg-mobile/trunk/templates/release_checklist.html.erb').read
  end

  def initialize(version, options = {})
    @version = version
    local_template = options.fetch(:template, false)
    @template = self.class.template
    @release_date = options.fetch(:release_date, "[date]")
    @mobile_version = options.fetch(:mobile_version, "XX.X")
    @is_unscheduled = options.fetch(:is_unscheduled, false)
    @unscheduled_message = options.fetch(:unscheduled_message, nil)
    @new_release_url="https://github.com/wordpress-mobile/gutenberg-mobile/releases/new?tag=v#{ @version }&amp;target=release/#{ @version }&amp;title=Release%20<% @version %>"

    @aztec_checklist = options.fetch(:aztec_checklist, '')
    @incoming_change_checklist = options.fetch(:incoming_change_checklist, '')

    super(@template)
  end

  def result
    super(binding)
  end

end

class AztecChecklist < Checklist
  def self.template
    if use_local_template?
      return File.read('./templates/aztec_checklist.html.erb')
    end
    begin 
      open('https://raw.githubusercontent.com/wordpress-mobile/release-toolkit-gutenberg-mobile/trunk/templates/aztec_checklist.html.erb').read
    rescue OpenURI::HTTPError => ex 
      abort_with "Error: #{ex.message}"
    end
  end

  def initialize()
    @template = self.class.template
    super(@template)
  end

  def result
    super(binding)
  end

  def self.render()
    new().result
  end
end

class IncomingChangeChecklist < Checklist
  def self.template
    if use_local_template?
      return File.read('./templates/incoming_change_checklist.html.erb')
    end
    open('https://raw.githubusercontent.com/wordpress-mobile/release-toolkit-gutenberg-mobile/trunk/templates/incoming_change_checklist.html.erb').read
  end

  def initialize(options = {})
    local_template = options.fetch(:template, false)
    @template = local_template ? File.read(local_template) : self.class.template
    @version = options.fetch(:version, 'XX.X')
    super(@template)
  end

  def result
    super(binding)
  end

  def self.render(options)
    new(options).result
  end
end

# Parse cli options
options = {}

option_parser = OptionParser.new do |opts|
  opts.banner = "Usage: release_checklist.rb version [options]"
  opts.on '-d', '--release-date RELEASE_DATE', 'Release date' do |d|
    options[:release_date] = d
  end

  opts.on '-v', '--mobile-version', 'Mobile host version (Only used for unscheduled releases)' do |v|
    options[:mobile_version] = v
  end

  opts.on '-o', '--output OUTPUT', 'Output file' do |o|
    options[:output] = o
  end

  opts.on '-m', '--message MESSAGE', 'Unscheduled release message' do |m|
    options[:unscheduled_message] = m
  end
end
option_parser.parse!

@version = ARGV[0]

# Helper methods
def confirm?(prompt)
  STDERR.print(prompt)
  STDIN.gets.strip.downcase == 'y'
end

def is_scheduled?
  @version.split('.')[-1].to_i.zero?
end

def can_open_editor?
  ENV["EDITOR"] && !ENV["EDITOR"].empty? && STDOUT.tty?
end



def skip_aztec_checklist?
  #TODO: Add logic to skip Aztec checklist
  false
end

# Validate parameters
abort_with "Valid version is required ( X.XX.X format )" unless @version.match?(/^\d+\.\d+\.\d+$/)

# Start prompts for an unscheduled release
unless is_scheduled?
  abort_with "Please verify the release version is correct." unless confirm? "Is this an unscheduled release? (y/n) "

  options[:is_unscheduled] = true

  # Check if we should add a message to the checklist
  if can_open_editor? && options[:unscheduled_message].nil?
     if confirm? "Do you want to add a message to the checklist? (y/n) "
        at_exit { @message_fd&.unlink }
        @message_fd = Tempfile.new('release_checklist_message')

        system(ENV["EDITOR"], @message_fd.path)

        @message_fd.rewind
        options[:unscheduled_message] = @message_fd.read
        say "Message added to checklist: #{options[:unscheduled_message]}"
        @message_fd.close
     end
  end
end

# Add optional steps to the checklist
options[:aztec_checklist] = AztecChecklist.render() unless skip_aztec_checklist?

checklist = ReleaseChecklist.new(@version, options).result

if options[:output]
  exit_with "Checklist written to #{options[:output]} 🚀" if File.write(options[:output], checklist)
  abort_with "Error writing checklist to #{options[:output]}"
end

say "Checklist complete 🚀" unless STDOUT.tty?
print checklist
