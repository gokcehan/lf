local stringify = pandoc.utils.stringify
local List = pandoc.List

-- We use bold paragraphs for flags (e.g. **-help**)
local function is_term(b)
	return b.tag == "Para" and b.content[1] and b.content[1].tag == "Strong"
end

-- GitHub-flavored markdown lacks native definition lists (i.e. ":" syntax).
-- Therefore, we generate them on the fly to get a better-looking "OPTIONS"
-- section in man pages and plain-text help without changing the source file.
function Blocks(blocks)
	local out = List()
	local i = 1
	local in_opts, lvl = false, nil

	while blocks[i] do
		local b = blocks[i]
		i = i + 1

		if b.tag == "Header" then
			-- Apply definition lists inside "OPTIONS" only
			if stringify(b.content) == "OPTIONS" then
				in_opts, lvl = true, b.level
			else
				in_opts = in_opts and b.level > lvl
			end
			out:insert(b)
		elseif in_opts and is_term(b) then
			-- Collect definition for current term
			local def = List()
			while blocks[i] and blocks[i].tag ~= "Header" and not is_term(blocks[i]) do
				def:insert(blocks[i])
				i = i + 1
			end
			out:insert(pandoc.DefinitionList({ { b.content, { def } } }))
		else
			-- Pass through unchanged
			out:insert(b)
		end
	end

	return out
end
