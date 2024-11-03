function getQueryParams() {
    // for (const [key, value] of mySearchParams) {
    //     console.log(key, value);
    // }

    if (window.location.search) {
        const queryString = window.location.search;
        console.log(queryString);
        if (queryString === '?page=add') {
            console.log('1');
            window.location = '/page-add.html', '_self';
            console.log('2');
        }
    }
}

function openPage(pageName) {
    window.location = pageName, '_self';
}

function getNavContent() {
    let navContent ='';
    navContent += "<nav class=\"navbar navbar-expand-lg bg-body-tertiary\">";
    navContent +="    <div class=\"container-fluid\">";
    navContent +="    <a class=\"navbar-brand\" href=\"#\" onClick=\"openPage('/index.html')\">Home</a>";
    navContent +="        <button class=\"navbar-toggler\" type=\"button\" data-bs-toggle=\"collapse\" data-bs-target=\"#navbarSupportedContent\" aria-controls=\"navbarSupportedContent\" aria-expanded=\"false\" aria-label=\"Toggle navigation\">";
    navContent +="        <span class=\"navbar-toggler-icon\"></span>";
    navContent +="</button>";
    navContent +="<div class=\"collapse navbar-collapse\" id=\"navbarSupportedContent\">";
    navContent +="        <ul class=\"navbar-nav me-auto mb-2 mb-lg-0\">";
    navContent +="        <li class=\"nav-item\">";
    navContent +="          <a class=\"nav-link\" href=\"#\" aria-current=\"page\" onclick=\"openPage('/pages.html')\">Pages";
    navContent +="                    </a>";
    navContent +="        </li>";
    navContent +="      <li class=\"nav-item\">";
    navContent +="          <a class=\"nav-link\" href=\"#\" onclick=\"openPage('/pictures.html')\">Pictures</a>";
    navContent +="  </li>";
    navContent +="  <li class=\"nav-item\">";
    navContent +="    <a class=\"nav-link\" href=\"#\" onclick=\"openPage('/links.html')\">Links</a>";
    navContent +="  </li>";
    navContent +="  <li class=\"nav-item\">";
    navContent +="      <a class=\"nav-link\" href=\"#\" onclick=\"openPage('/abbreviations.html')\">Abbreviations</a>";
    navContent +="  </li>";
    navContent +="";

    navContent +="  <li class=\"nav-item dropdown\">";
    navContent +="      <a class=\"nav-link dropdown-toggle\" href=\"#\" role=\"button\" data-bs-toggle=\"dropdown\" aria-expanded=\"false\">Add</a>";
    navContent +="      <ul class=\"dropdown-menu\">";
    navContent +="          <li><a class=\"dropdown-item\" href=\"#\" onClick=\"openPage('/page-add.html')\">Page</a></li>";
    navContent +="          <li><a class=\"dropdown-item\" href=\"#\" onClick=\"openPage('/link-add.html')\">Link</a></li>";
    navContent +="          <li><a class=\"dropdown-item\" href=\"#\" onClick=\"openPage('/abbreviation-add.html')\">Abbreviation</a></li>";
    navContent +="      </ul>";
    navContent +="  </li>";
    navContent +="</ul>";
    navContent +="<form class=\"d-flex\" role=\"search\">";
    navContent +="  <input class=\"form-control me-2\" type=\"search\" placeholder=\"Search\" aria-label=\"Search\"/>";
    navContent +="  <button class=\"btn btn-outline-success\" type=\"submit\">Search</button>";
    navContent +="</form>";
    navContent +="</div>";
    navContent +="</div>";
    navContent +="</nav>";
    return navContent;
}

