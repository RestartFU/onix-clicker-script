-- AutoClicker.lua

name = "Auto Clicker"
description = "Toggle the external clicker and update CPS"

ToggleKey = 0x77 -- F8

CpsValue = 12
LastCpsValue = CpsValue
LastCpsSentAt = 0
LastCpsChangedAt = 0
CpsDebounceSeconds = 0.75
ClickerEnabled = false
StateDir = "C:\\Users\\danick\\AppData\\Local\\Packages\\Microsoft.MinecraftUWP_8wekyb3d8bbwe\\RoamingState\\OnixClient\\Scripts\\AutoComplete\\clicker"
StateFile = StateDir .. "\\state.json"
FallbackStateDir = "clicker"
FallbackStateFile = FallbackStateDir .. "/state.json"
ShowOverlay = true

positionX = 10
positionY = 100
sizeX = 60
sizeY = 7


function render()
    if not ShowOverlay then
        return
    end
    local state = "Disabled"
    if ClickerEnabled then
        state = "Enabled"
    end
    local text = "Clicker: [CPS: " .. tostring(CpsValue) .. ", " .. state .. "]"

    local font = gui.font()
    sizeX = font.width(text) + 0.2 + paddingSetting.value * 2
    sizeY = font.height * 1.2 + paddingSetting.value * 2

    gfx.color(bgColorSetting.value.r, bgColorSetting.value.g, bgColorSetting.value.b, bgColorSetting.value.a)
    gfx.rect(
        0, 0,
        font.width(text) + 0.2 + paddingSetting.value * 2,
        font.height * 1.2 + paddingSetting.value * 2
    )

    gfx.color(textColorSetting.value.r, textColorSetting.value.g, textColorSetting.value.b, textColorSetting.value.a)
    gfx.text(paddingSetting.value, paddingSetting.value, text)
end

local function ensureStateDir(dir)
    if not fs.isdir(dir) then
        fs.mkdir(dir)
    end
end

local function writeState()
    ensureStateDir(StateDir)
    local enabledText = "false"
    if ClickerEnabled then
        enabledText = "true"
    end
    local payload = string.format([[{"enabled":%s,"cps":%d}]], enabledText, CpsValue)
    local function writeWithFs(path)
        if fs.writefile then
            local ok = pcall(fs.writefile, path, payload)
            if ok then
                return true
            end
        end
        return false
    end

    local function writeWithIo(path)
        if not io or not io.open then
            return false
        end
        local f = io.open(path, "w")
        if not f then
            return false
        end
        f:write(payload)
        f:flush()
        f:close()
        return true
    end

    local function writeWithHandle(path)
        local ok, f = pcall(fs.open, path, "w")
        if not ok then
            return false
        end
        if not f then
            return false
        end
        f:write(payload)
        f:flush()
        f:close()
        return true
    end

    ensureStateDir(FallbackStateDir)

    if writeWithFs(StateFile) then
        return
    elseif writeWithFs(FallbackStateFile) then
        return
    elseif writeWithIo(StateFile) then
        return
    elseif writeWithIo(FallbackStateFile) then
        return
    elseif writeWithHandle(StateFile) then
        return
    elseif writeWithHandle(FallbackStateFile) then
        return
    end
end

function ToggleClicker()
    ClickerEnabled = not ClickerEnabled
    writeState()

    if gui.screen() == "hud" or gui.screen() == "" then
        gui.setGrab(false)
    end
end

function ApplyCps()
    if CpsValue <= 0 then
        return
    end
    writeState()

    if gui.screen() == "hud" or gui.screen() == "" then
        gui.setGrab(false)
    end
end

client.settings.addCategory("Clicker Control")
client.settings.addKeybind("Toggle Key", "ToggleKey")
client.settings.addInt("CPS", "CpsValue", 1, 30)
client.settings.addBool("Show Overlay", "ShowOverlay")

textColorSetting = client.settings.addNamelessColor("Text Color", { 255, 255, 255, 255 })
bgColorSetting = client.settings.addNamelessColor("Background Color", { 51, 51, 51, 100 })
paddingSetting = client.settings.addNamelessFloat("Padding", 0, 10, 1)

client.settings.addFunction("Toggle Clicker", "ToggleClicker", "Toggle")
client.settings.stopCategory()

event.listen("KeyboardInput", function(key, down)
    if down and key == ToggleKey then
        ToggleClicker()
    end
end)

event.listen("ConfigurationLoaded", function()
    writeState()
end)

event.listen("Tick", function()
    local now = os.clock()
    if CpsValue ~= LastCpsValue then
        LastCpsValue = CpsValue
        LastCpsChangedAt = now
    end

    if LastCpsChangedAt > 0 and (now - LastCpsChangedAt) >= CpsDebounceSeconds then
        LastCpsChangedAt = 0
        LastCpsSentAt = now
        ApplyCps()
    end
end)
