# goback - Simple Incremental Backups

goback allows incremental file system backups that always keep the last version complete.

For every backup a directory (named after the current timestamp) is created and the last backup is used as a reference. The folder structure
of the target directory looks like this:

    backup                  - The target directory
    +--config.goback        - Metadata for the backup
    +--2020-04-02/[...]     - The first backup
    +--2020-04-02.goback    - Change detection data for the first backup
    +--2020-04-03/[...]     - The second backup
    +--2020-04-03.goback    - Change detection data for the second backup
    +--2020-04-04/[...]     - The third backup
    +--2020-04-04.goback    - Change detection data for the third backup
    [...]

The config.goback file stores the configuration values for the last backup, the other .goback-files contain the change-detection data for the
backups and the folders contain the backup data.
The data inside the .goback-files is stored as JSON.

## Usage

Initial backup:

    goback [-type daily] [-level 3] -source SOURCE TARGET

Subsequent backups:

    goback [-type daily] [-level 3] [-source SOURCE] TARGET

The arguments from the first backup will be saved inside the configuration (except level) and can be overwritten by providing different arguments for the next backup.

### Example

Backup directory `/home/user/data` to `/mnt/backup/userdata` and only create a new backup folder once every day:

    goback -type daily -source /home/user/data/ /mnt/backup/userdata/

The type and source values should be saved in `/mnt/backup/userdata/config.goback` and thus must not be specified on the next backup unless they changed:

    goback /mnt/backup/userdata/

## Motivation

I regularly backup my photo collection, which is now over 4TB, and I want to always be able to see the full directory structure for the latest backup.

Before writing goback, I used rsync to create backups with a reference directory that created hard links for files that did not change.
This had the advantage that the file system contained a complete representation for every backup while only using space for changed files.

The rsync-approach had two disadvantages for me:

 1. I could not easily see what changed between to backups
 2. When changing the filesystem or moving the whole backup to another disk rsync's change detection broke

goback addresses both problems:

 1. Every backup (apart from the last one) contains only the files that changed to the next one
 2. Since all metadata is stored in files and not attributes and there is a way to recover, old backups can be easily moved or deleted

## TODOs

My current plan is to add the following:

- Increase test coverage
- Write tests for the following cases
  - Multiple backups into the same timestamped backup
  - Changing configuration values after first backup
- Show progress speed in xB/s
- Show backup summary at the end
- Use github CI integration to check compilation, create binaries and report test coverage
