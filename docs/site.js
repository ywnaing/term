const downloadTabs = document.querySelectorAll("[data-download-tab]");
const downloadPanels = document.querySelectorAll("[data-download-panel]");

function selectDownloadTab(name) {
  downloadTabs.forEach((tab) => {
    const isActive = tab.dataset.downloadTab === name;
    tab.classList.toggle("active", isActive);
    tab.setAttribute("aria-selected", String(isActive));
  });

  downloadPanels.forEach((panel) => {
    panel.hidden = panel.dataset.downloadPanel !== name;
  });
}

downloadTabs.forEach((tab) => {
  tab.addEventListener("click", () => selectDownloadTab(tab.dataset.downloadTab));
});
