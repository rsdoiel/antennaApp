
# Development Notes

## Text Casting

- Textcasting website, <https://textcasting.org>
- Uses "source" namespace in RSS, see <https://source.scripting.com>
  - also see <https://scripting.com/?tab=links/rss.xml>
  - and <http://scripting.com/2022/07/19/152235.html?title=devNotesForMarkdownInFeeds>
  - and <http://xmlviewer.scripting.com/?url=https://wordland.social/scripting/237777565/rss.xml>
- GitHub issue
  - <https://github.com/scripting/sourceNamespaceComments/issues/3>
  - I need to ask the question to understand how I approach it.

## Upgrading schema

I've alter my SQLite schema several times. I need to implement an orderly upgrade process. I'm thinking the command should be something like `antenna upgrade` to update the schema in all the feed collections.

Since upgrades will accumulate I think they should go in a sql_upgrade.go instead of sql_stmts.go

Here's an outline of what is needed at the SQL level.

~~~sql
-- Check if the sourceMarkdown column exists
SELECT COUNT(*) AS cnt
FROM pragma_table_info('items')
WHERE name = 'sourceMarkdown';

~~~

If this returns zero rows then I would execute the

~~~sql
-- add column to table
ALTER TABLE items ADD COLUMN sourceMarkdown TEXT;

~~~

The ugprade process should support adding the postPath and sourceMarkdown columns in the iterms table.

## Namespace in RSS 

Since I am using RSS 2.0 as the guide to how I'm implementing the items schema I need to think more about how I want to integrate namespace more generally. I'm beinging to think the source namespace should be a map[string]interface{} like the dcExt. This would let me grow the support without changing the SQL schema as I add fields.

In this approach only the core fields would be columns and each complex relations would remain as they are as either an object or list. 

