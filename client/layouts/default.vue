<template>
    <div class="drawer drawer-mobile">
        <input id="sidebar" type="checkbox" class="drawer-toggle" />
        <div class="drawer-content flex flex-col">
            <!-- Navbar -->
            <AppNavbar />

            <!-- Page content here -->
            <div class="p-4">
                <slot></slot>
            </div>
        </div>

        <div class="drawer-side">
            <label for="sidebar" class="drawer-overlay"></label>
            <AppSidebar />
        </div>
    </div>
</template>
