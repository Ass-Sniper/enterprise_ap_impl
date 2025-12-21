package main

const (
	// SuccessPageTemplate 成功管理页面
	SuccessPageTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <script src="https://cdn.tailwindcss.com"></script>
    %s
</head>
<body class="bg-gray-50 flex items-center justify-center min-h-screen">
    <div class="max-w-md w-full bg-white shadow-lg rounded-xl p-8 text-center">
        <div class="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-green-100 mb-6">
            <svg class="h-10 w-10 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
        </div>
        <h2 class="text-2xl font-bold text-gray-800 mb-2">认证成功</h2>
        <div class="bg-blue-50 rounded-lg p-4 mb-8 text-left">
            <div class="flex justify-between text-sm"><span class="text-blue-700">当前用户:</span><span class="font-bold">%s</span></div>
        </div>
        <form action="%s" method="post" id="authForm" class="hidden">
            <input type="hidden" name="username" value="%s"><input type="hidden" name="token" value="%s">
        </form>
        <div class="space-y-4">
            <p class="auto-submit-msg text-xs text-gray-400 italic">正在自动同步准入状态...</p>
            <form action="%s" method="POST">
                <input type="hidden" name="username" value="%s">
                <button type="submit" class="active-scale w-full py-3 rounded-lg bg-red-600 text-white text-sm font-medium">解除认证并断开网络</button>
            </form>
        </div>
    </div>
    %s
</body>
</html>`

	// FailurePageTemplate 失败提示页面
	FailurePageTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <script src="https://cdn.tailwindcss.com"></script>
    %s
</head>
<body class="bg-gray-50 flex items-center justify-center min-h-screen">
    <div class="max-w-md w-full bg-white shadow-lg rounded-xl p-8 text-center shake">
        <div class="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-red-100 mb-6">
            <svg class="h-10 w-10 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
        </div>
        <h2 class="text-2xl font-bold text-gray-800 mb-2">认证失败</h2>
        <p class="text-red-500 mb-8">%s</p>
        <a href="/portal" class="block w-full py-3 rounded-lg bg-gray-800 text-white font-medium">立即返回</a>
        <p class="mt-4 text-xs text-gray-400"><span id="countdown">5</span> 秒后自动跳转...</p>
    </div>
    %s
</body>
</html>`
)
