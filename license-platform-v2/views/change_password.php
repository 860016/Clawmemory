<div class="max-w-md mx-auto mt-8 bg-white rounded-lg shadow p-6">
    <h2 class="text-xl font-bold text-gray-900 mb-6">修改密码</h2>

    <?php if (!empty($success)): ?>
    <div class="mb-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded"><?= htmlspecialchars($success) ?></div>
    <?php endif; ?>

    <?php if (!empty($error)): ?>
    <div class="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded"><?= htmlspecialchars($error) ?></div>
    <?php endif; ?>

    <form method="POST" action="/admin/change-password" class="space-y-4">
        <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">旧密码</label>
            <input type="password" name="old_password" required
                class="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 border p-2">
        </div>
        <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">新密码（至少6位）</label>
            <input type="password" name="new_password" required minlength="6"
                class="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 border p-2">
        </div>
        <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">确认新密码</label>
            <input type="password" name="confirm_password" required minlength="6"
                class="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 border p-2">
        </div>
        <button type="submit"
            class="w-full py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700">
            确认修改
        </button>
    </form>
    <div class="mt-4 text-center">
        <a href="/admin/dashboard" class="text-sm text-gray-500 hover:text-indigo-600">返回仪表盘</a>
    </div>
</div>
