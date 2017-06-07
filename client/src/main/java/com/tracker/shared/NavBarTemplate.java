package com.tracker.shared;

import java.util.Map;

import com.googlecode.jatl.client.HtmlWriter;
import com.googlecode.jatl.client.MarkupBuilder.TagClosingPolicy;

public class NavBarTemplate extends BaseTemplates{
  public HtmlWriter renderNavBar(final Map<String, Object> map) {
    return new HtmlWriter() {
      @Override
      protected void build() {
        indent(indentOff);
        renderNavBar(this, map);
      }};
  }

  private void renderNavBar(HtmlWriter html, final Map<String, Object> map) {
		html.start("nav", TagClosingPolicy.PAIR)
      .classAttr("navbar navbar-toggleable-md fixed-top navbar-light")
      .style("background-color: #dddddd;");

    html.a().classAttr("navbar-brand").href("#").text("Dashboard").end();

    html.div().classAttr("collapse navbar-collapse");
   
    html.ul().classAttr("navbar-nav mr-auto");
    for (Map<String, Object> item : getList(map, "items")) {
      html.li().classAttr("nav-item");
      html.a().classAttr("nav-link").href(getString(item, "href"))
          .text(getString(item, "name")).end();
      html.end(); // li
    }
    html.end(); // ul

    html.ul().classAttr("navbar-nav navbar-right");
      html.li().classAttr("nav-item");
      html.a().classAttr("btn btn-outline-primary btn-logout").href("#");
        html.span().classAttr("user-username").text("smalex").end();
        html.raw(" &nbsp;");
        html.i().classAttr("fa fa-sign-out fa-lg").text("").end();
      html.end(); // a
      html.end(); // li
    html.end(); // ul

    html.end(); // navbar-collapse

    html.end(); // nav
  }
}
