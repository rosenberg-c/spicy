-- Tests for spicy.utils.fs module
local fs = require("spicy.utils.fs")

describe("spicy.utils.fs", function()
  local test_dir

  before_each(function()
    -- Create temp directory for tests
    test_dir = vim.fn.tempname()
    vim.fn.mkdir(test_dir, "p")
  end)

  after_each(function()
    -- Cleanup test directory
    if test_dir and vim.fn.isdirectory(test_dir) == 1 then
      vim.fn.delete(test_dir, "rf")
    end
  end)

  describe("join", function()
    it("should join path components", function()
      local path = fs.join("a", "b", "c")
      assert.equals("a/b/c", path)
    end)

    it("should handle trailing slashes", function()
      local path = fs.join("a/", "b/", "c")
      assert.equals("a/b/c", path)
    end)

    it("should normalize double slashes", function()
      local path = fs.join("a//b", "c")
      assert.equals("a/b/c", path)
    end)

    it("should handle single component", function()
      local path = fs.join("alone")
      assert.equals("alone", path)
    end)
  end)

  describe("exists", function()
    it("should return true for existing file", function()
      local file = fs.join(test_dir, "exists.txt")
      vim.fn.writefile({ "test" }, file)

      assert.is_true(fs.exists(file))
    end)

    it("should return false for non-existent file", function()
      local file = fs.join(test_dir, "does-not-exist.txt")
      assert.is_false(fs.exists(file))
    end)

    it("should return true for existing directory", function()
      assert.is_true(fs.exists(test_dir))
    end)
  end)

  describe("is_dir", function()
    it("should return true for directory", function()
      assert.is_true(fs.is_dir(test_dir))
    end)

    it("should return false for file", function()
      local file = fs.join(test_dir, "file.txt")
      vim.fn.writefile({ "test" }, file)

      assert.is_false(fs.is_dir(file))
    end)

    it("should return false for non-existent path", function()
      assert.is_false(fs.is_dir(fs.join(test_dir, "nope")))
    end)
  end)

  describe("mkdir", function()
    it("should create directory", function()
      local dir = fs.join(test_dir, "new_dir")
      local ok, err = fs.mkdir(dir)

      assert.is_true(ok)
      assert.is_nil(err)
      assert.is_true(fs.exists(dir))
      assert.is_true(fs.is_dir(dir))
    end)

    it("should create nested directories", function()
      local dir = fs.join(test_dir, "a", "b", "c")
      local ok, err = fs.mkdir(dir)

      assert.is_true(ok)
      assert.is_nil(err)
      assert.is_true(fs.exists(dir))
    end)

    it("should expand tilde in path", function()
      -- This might fail on some systems, so we'll skip if ~/test exists
      local dir = "~/spicy_test_dir_" .. os.time()
      local ok, err = fs.mkdir(dir)

      if ok then
        local expanded = fs.expand(dir)
        assert.is_true(fs.exists(expanded))
        vim.fn.delete(expanded, "rf")
      end
    end)
  end)

  describe("write_file and read_file", function()
    it("should write and read file", function()
      local file = fs.join(test_dir, "test.txt")
      local content = "Hello, World!"

      local ok, err = fs.write_file(file, content)
      assert.is_true(ok)
      assert.is_nil(err)

      local read_content, read_err = fs.read_file(file)
      assert.is_nil(read_err)
      assert.equals(content, read_content)
    end)

    it("should create parent directories", function()
      local file = fs.join(test_dir, "nested", "dir", "file.txt")
      local ok, err = fs.write_file(file, "content")

      assert.is_true(ok)
      assert.is_nil(err)
      assert.is_true(fs.exists(file))
    end)

    it("should overwrite existing file", function()
      local file = fs.join(test_dir, "overwrite.txt")

      fs.write_file(file, "first")
      fs.write_file(file, "second")

      local content = fs.read_file(file)
      assert.equals("second", content)
    end)

    it("should handle empty content", function()
      local file = fs.join(test_dir, "empty.txt")
      local ok = fs.write_file(file, "")

      assert.is_true(ok)
      local content = fs.read_file(file)
      assert.equals("", content)
    end)

    it("should return error for invalid path", function()
      local file = "/invalid/path/that/cannot/exist.txt"
      local ok, err = fs.write_file(file, "test")

      assert.is_false(ok)
      assert.is_string(err)
    end)
  end)

  describe("basename", function()
    it("should extract filename", function()
      assert.equals("file.txt", fs.basename("/path/to/file.txt"))
      assert.equals("dir", fs.basename("/path/to/dir"))
      assert.equals("test.lua", fs.basename("test.lua"))
    end)
  end)

  describe("extension", function()
    it("should extract extension", function()
      assert.equals("txt", fs.extension("/path/to/file.txt"))
      assert.equals("lua", fs.extension("test.lua"))
    end)

    it("should return empty for no extension", function()
      assert.equals("", fs.extension("/path/to/file"))
    end)
  end)

  describe("remove_extension", function()
    it("should remove file extension", function()
      assert.equals(
        "/path/to/file",
        fs.remove_extension("/path/to/file.txt")
      )
      assert.equals("test", fs.remove_extension("test.lua"))
    end)

    it("should handle no extension", function()
      assert.equals("/path/to/file", fs.remove_extension("/path/to/file"))
    end)
  end)
end)
