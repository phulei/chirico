import java.sql.* ;


public class Prog1 {


    public static void main(String[] args) {


	try {

	    Class.forName("com.mysql.jdbc.Driver").newInstance();
	    Connection conn = DriverManager.getConnection("jdbc:mysql://localhost/scrapdatabase","prog1","s3cr3tpass45");

	    Statement stmt = null;
	    ResultSet rs = null;

	    stmt = conn.createStatement();
	    rs = stmt.executeQuery("call junk(1)");

         while (rs.next(  ))
                System.out.println(rs.getString(1));
            rs.close(  );
            stmt.close(  );
            conn.close(  );

	} catch (Exception ex) {
	    // handle any errors
	    System.out.println(ex);
	}






    }

}